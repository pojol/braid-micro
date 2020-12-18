package mailboxnsq

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/internal/buffer"
	"github.com/pojol/braid/module/mailbox"
)

type procMsg struct {
	msg     *mailbox.Message
	channel string
}

type procMailbox struct {
	subscribers map[string]*procSubscriber
	recvBuff    *buffer.Unbounded

	exitChan chan int
	guard    sync.Mutex
}

type procSubscriber struct {
	// channel 信道名
	channel string

	// 信道的模式
	mode string

	// 这个信道上的消费者
	consumers []mailbox.IConsumer
	lock      sync.RWMutex
}

type procConsumer struct {
	guard  sync.Mutex
	handle mailbox.HandlerFunc
	exitCh *braidsync.Switch
}

func (c *procConsumer) PutMsg(msg *mailbox.Message) error {

	if c.handle != nil {
		c.handle(*msg)
	}

	return nil
}

func (c *procConsumer) Exit() {
	c.exitCh.Open()
}

func (c *procConsumer) IsExited() bool {
	return c.exitCh.HasOpend()
}

func (c *procConsumer) OnArrived(handler mailbox.HandlerFunc) error {

	c.guard.Lock()
	c.handle = handler
	c.guard.Unlock()

	return nil
}

func (ns *procSubscriber) Competition() (mailbox.IConsumer, error) {

	if ns.mode == mailbox.Shared {
		return nil, errors.New("channel mode mutex")
	}

	ns.lock.Lock()
	defer ns.lock.Unlock()

	ns.mode = mailbox.Competition
	competition := &procConsumer{
		exitCh: braidsync.NewSwitch(),
	}
	ns.consumers = append(ns.consumers, competition)

	return competition, nil
}

func (ns *procSubscriber) Shared() (mailbox.IConsumer, error) {

	if ns.mode == mailbox.Competition {
		return nil, errors.New("channel mode mutex")
	}

	ns.lock.Lock()
	defer ns.lock.Unlock()

	ns.mode = mailbox.Shared
	shared := &procConsumer{
		exitCh: braidsync.NewSwitch(),
	}
	ns.consumers = append(ns.consumers, shared)

	return shared, nil
}

func (pmb *procMailbox) router() {

	for {
		select {
		case msg := <-pmb.recvBuff.Get():
			pmsg := msg.(*procMsg)
			pmb.recvBuff.Load()

			s, ok := pmb.subscribers[pmsg.channel]
			if ok {

				if s.mode == mailbox.Shared {

					for k := range s.consumers {
						if !s.consumers[k].IsExited() {
							s.consumers[k].PutMsg(pmsg.msg)
						}
					}

				} else if s.mode == mailbox.Competition {
					s.consumers[rand.Intn(len(s.consumers))].PutMsg(pmsg.msg)
				}
			}

		}
	}

}

func (pmb *procMailbox) pub(topic string, msg *mailbox.Message) {

	pmsg := &procMsg{
		msg:     msg,
		channel: topic,
	}

	pmb.recvBuff.Put(pmsg)

	/*
		select {
		case pmb.recvBuff <- pmsg:
		case <-pmb.exitChan:
			// return err
		}
	*/
}

func (pmb *procMailbox) sub(topic string) mailbox.ISubscriber {

	pmb.guard.Lock()
	defer pmb.guard.Unlock()

	s, ok := pmb.subscribers[topic]
	if ok {
		return s
	}

	s = &procSubscriber{
		channel: topic,
		mode:    mailbox.Undecided,
	}
	pmb.subscribers[topic] = s

	return s
}
