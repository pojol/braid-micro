package mailboxnsq

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/internal/braidsync"
	"github.com/pojol/braid-go/module/mailbox"
)

type mailboxTopic struct {
	Name    string
	mailbox *nsqMailbox
	scope   mailbox.ScopeTy

	msgch    chan *mailbox.Message
	exitFlag int32
	producer []*nsq.Producer

	waitGroup braidsync.WaitGroupWrapper

	startChan         chan int
	exitChan          chan int
	channelUpdateChan chan int

	sync.RWMutex

	channelMap map[string]*mailboxChannel
}

func newTopic(name string, scope mailbox.ScopeTy, n *nsqMailbox) *mailboxTopic {

	topic := &mailboxTopic{
		Name:              name,
		mailbox:           n,
		scope:             scope,
		startChan:         make(chan int, 1),
		exitChan:          make(chan int),
		channelUpdateChan: make(chan int),
		msgch:             make(chan *mailbox.Message, 4096),
		channelMap:        make(map[string]*mailboxChannel),
	}

	if scope == mailbox.ScopeCluster {
		cps := make([]*nsq.Producer, 0, len(n.parm.NsqdAddress))
		var err error
		var cp *nsq.Producer

		for _, addr := range n.parm.LookupAddress {

			url := fmt.Sprintf("http://%s/topic/create?topic=%s",
				addr,
				name,
			)
			req, err := http.NewRequest("POST", url, nil)
			if err != nil {
				n.log.Warn(err.Error())
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
				n.log.Errorf("Channel new nsq producer err %v", err.Error())
				continue
			}

			if err = cp.Ping(); err != nil {
				n.log.Errorf("Channel nsq producer ping err %v", err.Error())
				continue
			}

			cps = append(cps, cp)

			url := fmt.Sprintf("http://%s/topic/create?topic=%s", addr, name)
			resp, err := http.Post(url, "application/json", nil)
			if err != nil {
				n.log.Warn(err.Error())
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
	}

	topic.waitGroup.Wrap(topic.loop)

	return topic

}

func (t *mailboxTopic) start() {
	select {
	case t.startChan <- 1:
	default:
	}
}

func (t *mailboxTopic) Sub(name string) mailbox.IChannel {

	t.Lock()
	c, isNew := t.getOrCreateChannel(name, t.scope)
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

func (t *mailboxTopic) getOrCreateChannel(name string, scope mailbox.ScopeTy) (mailbox.IChannel, bool) {

	channel, ok := t.channelMap[name]
	if !ok {
		channel = newChannel(t.Name, name, scope, t.mailbox)
		t.channelMap[name] = channel

		t.mailbox.log.Infof("Topic %v new channel %v", t.Name, name)
		return channel, true
	}

	return channel, false

}

func (t *mailboxTopic) put(msg *mailbox.Message) error {
	select {
	case t.msgch <- msg:
	default:
		return fmt.Errorf("the mailbox topic %v queue is full!", t.Name)
	}

	return nil
}

func (t *mailboxTopic) loop() {
	var msg *mailbox.Message
	var chans []*mailboxChannel
	var msgChan chan *mailbox.Message

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
}

func (t *mailboxTopic) Pub(msg *mailbox.Message) error {
	t.RLock()
	defer t.RUnlock()

	if atomic.LoadInt32(&t.exitFlag) == 1 {
		return errors.New("exiting")
	}

	if t.scope == mailbox.ScopeProc {
		err := t.put(msg)
		if err != nil {
			return err
		}
	} else {
		t.producer[rand.Intn(len(t.producer))].Publish(t.Name, msg.Body)
	}

	return nil
}

func (t *mailboxTopic) Exit() error {

	if !atomic.CompareAndSwapInt32(&t.exitFlag, 0, 1) {
		return errors.New("exiting")
	}

	close(t.exitChan)
	t.waitGroup.Wait()

	return nil
}
