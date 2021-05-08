package mailboxnsq

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/pojol/braid-go/internal/braidsync"
	"github.com/pojol/braid-go/module/mailbox"
)

type mailboxTopic struct {
	Name    string
	mailbox *nsqMailbox

	msgch    chan *mailbox.Message
	exitFlag int32

	waitGroup braidsync.WaitGroupWrapper

	startChan         chan int
	exitChan          chan int
	channelUpdateChan chan int

	sync.RWMutex

	channelMap map[string]*mailboxChannel
}

func newTopic(name string, n *nsqMailbox) *mailboxTopic {

	topic := &mailboxTopic{
		Name:              name,
		mailbox:           n,
		startChan:         make(chan int, 1),
		exitChan:          make(chan int),
		channelUpdateChan: make(chan int),
		msgch:             make(chan *mailbox.Message, 4096),
		channelMap:        make(map[string]*mailboxChannel),
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

func (t *mailboxTopic) Channel(name string, scope mailbox.ScopeTy) mailbox.IChannel {

	t.Lock()
	c, isNew := t.getOrCreateChannel(name, scope)
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
		return fmt.Errorf("the mailbox topic queue is full!")
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

	err := t.put(msg)
	if err != nil {
		return err
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
