package pubsubnsq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/internal/braidsync"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
)

type pubsubTopic struct {
	Name string
	ps   *nsqPubsub
	log  *blog.Logger

	msgch    chan *meta.Message
	exitFlag int32
	producer []*nsq.Producer

	waitGroup braidsync.WaitGroupWrapper

	startChan         chan int
	exitChan          chan int
	channelUpdateChan chan int

	sync.RWMutex

	channelMap map[string]*pubsubChannel
}

func newTopic(name string, n *nsqPubsub) *pubsubTopic {

	topic := &pubsubTopic{
		Name:              name,
		ps:                n,
		log:               n.log,
		startChan:         make(chan int, 1),
		exitChan:          make(chan int),
		channelUpdateChan: make(chan int),
		msgch:             make(chan *meta.Message, n.parm.ChannelLength),
		channelMap:        make(map[string]*pubsubChannel),
	}

	cps := make([]*nsq.Producer, 0, len(n.parm.NsqdAddress))
	var err error
	var cp *nsq.Producer

	for _, addr := range n.parm.LookupdAddress {

		url := fmt.Sprintf("http://%s/topic/create?topic=%s",
			addr,
			name,
		)
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			n.log.Warnf("post %v err %v", url, err.Error())
		}
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			if resp.StatusCode != http.StatusOK {
				n.log.Warnf("lookupd create topic request status err %v", resp.StatusCode)
			}
			ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		}

	}

	for k, addr := range n.parm.NsqdHttpAddress {
		cp, err = nsq.NewProducer(n.parm.NsqdAddress[k], nsq.NewConfig())
		if err != nil {
			n.log.Warnf("Channel new nsq producer err %v", err.Error())
			continue
		}

		if err = cp.Ping(); err != nil {
			n.log.Warnf("Channel nsq producer ping err %v addr %v", err.Error(), addr)
			continue
		}

		cps = append(cps, cp)

		url := fmt.Sprintf("http://%s/topic/create?topic=%s", addr, name)
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			n.log.Warnf("post url %v err %v", url, err.Error())
		}
		if resp != nil {
			if resp.StatusCode != http.StatusOK {
				n.log.Warnf("nsqd create topic request status err %v", resp.StatusCode)
			}

			ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}

	topic.producer = cps

	topic.waitGroup.Wrap(topic.loop)

	return topic

}

func (t *pubsubTopic) start() {
	select {
	case t.startChan <- 1:
	default:
	}
}

func (t *pubsubTopic) Sub(ctx context.Context, name string, opts ...interface{}) (module.IChannel, error) {

	t.Lock()
	c, isNew := t.getOrCreateChannel(name)
	t.Unlock()

	if isNew {
		// update loop state
		select {
		case t.channelUpdateChan <- 1:
		case <-t.exitChan:
		}
	}

	return c, nil
}

func (t *pubsubTopic) getOrCreateChannel(name string) (module.IChannel, bool) {

	channel, ok := t.channelMap[name]
	if !ok {
		channel = newChannel(t.Name, name, t.log, t)
		t.channelMap[name] = channel

		t.log.Infof("Topic %v new channel %v", t.Name, name)
		return channel, true
	}

	return channel, false

}

func (t *pubsubTopic) rmvChannel(name string) error {
	t.RLock()
	channel, ok := t.channelMap[name]
	t.RUnlock()
	if !ok {
		return fmt.Errorf("channel %v does not exist", name)
	}

	t.log.Infof("topic %v deleting channel %v", t.Name, name)
	err := channel.exit()
	if err != nil {
		return err
	}

	t.Lock()
	delete(t.channelMap, name)
	t.Unlock()

	select {
	case t.channelUpdateChan <- 1:
	case <-t.exitChan:
	}

	return nil
}

func (t *pubsubTopic) put(msg *meta.Message) error {
	select {
	case t.msgch <- msg:
	default:
		return fmt.Errorf("the pubsub topic %v queue is full", t.Name)
	}

	return nil
}

func (t *pubsubTopic) loop() {
	var msg *meta.Message
	var chans []*pubsubChannel
	var msgChan chan *meta.Message

	for {
		select {
		case <-t.channelUpdateChan:
			continue
		case <-t.exitChan:
			goto EXT
		case <-t.startChan:
		}
		break
	}

	t.RLock()
	for _, c := range t.channelMap {
		chans = append(chans, c)
	}
	t.RUnlock()
	if len(chans) > 0 {
		msgChan = t.msgch
	}

	for {
		select {
		case msg = <-msgChan:
		case <-t.channelUpdateChan:
			chans = chans[:0]
			t.RLock()
			for _, c := range t.channelMap {
				chans = append(chans, c)
			}
			t.RUnlock()
			if len(chans) == 0 {
				msgChan = nil
			} else {
				msgChan = t.msgch
			}
			continue
		case <-t.exitChan:
			goto EXT
		}

		for _, channel := range chans {
			channel.put(msg)
		}
	}

EXT:
	t.log.Infof("topic %v out of the loop", t.Name)
}

func (t *pubsubTopic) Pub(ctx context.Context, msg *meta.Message) error {
	t.RLock()
	defer t.RUnlock()

	if atomic.LoadInt32(&t.exitFlag) == 1 {
		return errors.New("exiting")
	}

	byt, _ := json.Marshal(msg)

	err := t.producer[rand.Intn(len(t.producer))].Publish(t.Name, byt)
	if err != nil {
		t.log.Warnf("topic %v publish err %v\n", t.Name, err.Error())
		return err
	}

	return nil
}

func (t *pubsubTopic) Close() error {
	return t.ps.rmvTopic(t.Name)
}

func (t *pubsubTopic) exit() error {

	if !atomic.CompareAndSwapInt32(&t.exitFlag, 0, 1) {
		return errors.New("exiting")
	}

	t.log.Infof("topic %v exiting", t.Name)

	close(t.exitChan)
	// 等待 loop 中止后处理余下逻辑
	t.waitGroup.Wait()

	t.Lock()
	for _, channel := range t.channelMap {
		delete(t.channelMap, channel.Name)
		channel.exit()
	}
	t.Unlock()

	return nil
}
