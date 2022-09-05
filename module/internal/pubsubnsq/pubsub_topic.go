package pubsubnsq

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/internal/braidsync"
	"github.com/pojol/braid-go/module/pubsub"
)

type pubsubTopic struct {
	Name string
	ps   *nsqPubsub

	msgch    chan *pubsub.Message
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
		startChan:         make(chan int, 1),
		exitChan:          make(chan int),
		channelUpdateChan: make(chan int),
		msgch:             make(chan *pubsub.Message, 4096),
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
			blog.Warnf(err.Error())
		}
		resp, _ := http.DefaultClient.Do(req)
		if resp != nil {
			if resp.StatusCode != http.StatusOK {
				blog.Warnf("lookupd create topic request status err %v", resp.StatusCode)
			}
			ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		}

	}

	for k, addr := range n.parm.NsqdHttpAddress {
		cp, err = nsq.NewProducer(n.parm.NsqdAddress[k], nsq.NewConfig())
		if err != nil {
			blog.Errf("Channel new nsq producer err %v", err.Error())
			continue
		}

		if err = cp.Ping(); err != nil {
			blog.Errf("Channel nsq producer ping err %v addr %v", err.Error(), addr)
			continue
		}

		cps = append(cps, cp)

		url := fmt.Sprintf("http://%s/topic/create?topic=%s", addr, name)
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			blog.Warnf("post url %v", err.Error())
		}
		if resp != nil {
			if resp.StatusCode != http.StatusOK {
				blog.Warnf("nsqd create topic request status err %v", resp.StatusCode)
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

func (t *pubsubTopic) Sub(name string) pubsub.IChannel {

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

	return c
}

func (t *pubsubTopic) getOrCreateChannel(name string) (pubsub.IChannel, bool) {

	channel, ok := t.channelMap[name]
	if !ok {
		channel = newChannel(t.Name, name, t.ps)
		t.channelMap[name] = channel

		//blog.Infof("Topic %v new channel %v", t.Name, name)
		return channel, true
	}

	return channel, false

}

func (t *pubsubTopic) RmvChannel(name string) error {
	t.RLock()
	channel, ok := t.channelMap[name]
	t.RUnlock()
	if !ok {
		return fmt.Errorf("channel %v does not exist", name)
	}

	//blog.Infof("topic %v deleting channel %v", t.Name, name)
	channel.Exit()

	t.Lock()
	delete(t.channelMap, name)
	t.Unlock()

	select {
	case t.channelUpdateChan <- 1:
	case <-t.exitChan:
	}

	return nil
}

func (t *pubsubTopic) loop() {
	var msg *pubsub.Message
	var chans []*pubsubChannel
	var msgChan chan *pubsub.Message

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
			channel.Put(msg)
		}
	}

EXT:
	//blog.Infof("topic %v out of the loop", t.Name)
}

func (t *pubsubTopic) Pub(msg *pubsub.Message) error {
	t.RLock()
	defer t.RUnlock()

	if atomic.LoadInt32(&t.exitFlag) == 1 {
		return errors.New("exiting")
	}

	err := t.producer[rand.Intn(len(t.producer))].Publish(t.Name, msg.Body)

	if err != nil {
		blog.Warnf("topic %v publish err %v", t.Name, err.Error())
		return err
	}

	return nil
}

func (t *pubsubTopic) Exit() error {

	if !atomic.CompareAndSwapInt32(&t.exitFlag, 0, 1) {
		return errors.New("exiting")
	}

	//blog.Infof("topic %v exiting", t.Name)

	close(t.exitChan)
	// 等待 loop 中止后处理余下逻辑
	t.waitGroup.Wait()

	t.Lock()
	for _, channel := range t.channelMap {
		delete(t.channelMap, channel.Name)
		channel.Exit()
	}
	t.Unlock()

	return nil
}
