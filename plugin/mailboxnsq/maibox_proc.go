package mailboxnsq

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/internal/buffer"
	"github.com/pojol/braid/module/mailbox"
)

type procMailbox struct {
	subscribers sync.Map
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
	buff   *buffer.Unbounded
	exitCh *braidsync.Switch
}

func (c *procConsumer) PutMsg(msg *mailbox.Message) error {
	c.buff.Put(msg)
	return nil
}

func (c *procConsumer) Exit() {
	c.exitCh.Open()
}

func (c *procConsumer) IsExited() bool {
	return c.exitCh.HasOpend()
}

func (c *procConsumer) OnArrived(handler mailbox.HandlerFunc) {
	go func() {
		for {
			select {
			case msg := <-c.buff.Get():
				handler(msg.(*mailbox.Message))
				c.buff.Load()
			case <-c.exitCh.Done():
			}

			if c.exitCh.HasOpend() {
				return
			}
		}
	}()
}

func (ns *procSubscriber) Competition() (mailbox.IConsumer, error) {

	if ns.mode == mailbox.Shared {
		return nil, errors.New("channel mode mutex")
	}

	ns.lock.Lock()
	defer ns.lock.Unlock()

	ns.mode = mailbox.Competition
	competition := &procConsumer{
		buff:   buffer.NewUnbounded(),
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
		buff:   buffer.NewUnbounded(),
		exitCh: braidsync.NewSwitch(),
	}
	ns.consumers = append(ns.consumers, shared)

	return shared, nil
}

func (pmb *procMailbox) pub(topic string, msg *mailbox.Message) {

	s, ok := pmb.subscribers.Load(topic)
	if !ok {
		return
	}

	sub, ok := s.(*procSubscriber)
	if !ok {
		return
	}

	sub.lock.RLock()
	defer sub.lock.RUnlock()

	if sub.mode == mailbox.Shared {
		for k := range sub.consumers {
			sub.consumers[k].PutMsg(msg)
		}
	} else if sub.mode == mailbox.Competition {
		sub.consumers[rand.Intn(len(sub.consumers))].PutMsg(msg)
	}

}

func (pmb *procMailbox) sub(topic string) mailbox.ISubscriber {

	s, ok := pmb.subscribers.Load(topic)
	if !ok { // create
		psub := &procSubscriber{
			channel: topic,
			mode:    mailbox.Undecided,
		}

		pmb.subscribers.Store(topic, psub)
		s = psub
	}

	return s.(mailbox.ISubscriber)
}
