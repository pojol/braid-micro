package mailboxnsq

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/module/mailbox"
)

type procMsg struct {
	msg     *mailbox.Message
	channel string
}

type procMailbox struct {
	subscribers map[string]*procSubscriber
	exitChan    chan int
	guard       sync.RWMutex
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
	exitCh *braidsync.Switch

	// buffer
	recvBuff chan mailbox.Message
	backlog  []mailbox.Message
	sync.Mutex
}

func (c *procConsumer) PutMsg(msg *mailbox.Message) {
	c.Lock()
	if len(c.backlog) == 0 {
		select {
		case c.recvBuff <- *msg:
			c.Unlock()
			return
		default:
		}
	}

	c.backlog = append(c.backlog, *msg)
	c.Unlock()
}

func (c *procConsumer) Exit() {
	c.exitCh.Open()
}

func (c *procConsumer) Done() {
	c.Lock()
	if len(c.backlog) > 0 {
		select {
		case c.recvBuff <- c.backlog[0]:
			c.backlog[0] = mailbox.Message{}
			c.backlog = c.backlog[1:]
		default:
		}
	}
	c.Unlock()
}

func (c *procConsumer) IsExited() bool {
	return c.exitCh.HasOpend()
}

func (c *procConsumer) OnArrived() <-chan mailbox.Message {
	return c.recvBuff
}

func (ns *procSubscriber) Competition() (mailbox.IConsumer, error) {

	if ns.mode == mailbox.Shared {
		return nil, errors.New("channel mode mutex")
	}

	ns.lock.Lock()
	defer ns.lock.Unlock()

	ns.mode = mailbox.Competition
	competition := &procConsumer{
		recvBuff: make(chan mailbox.Message, 1),
		exitCh:   braidsync.NewSwitch(),
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
		recvBuff: make(chan mailbox.Message, 1),
		exitCh:   braidsync.NewSwitch(),
	}
	ns.consumers = append(ns.consumers, shared)

	return shared, nil
}

func (pmb *procMailbox) pub(topic string, msg *mailbox.Message) {

	pmb.guard.RLock()
	defer pmb.guard.RUnlock()

	s, ok := pmb.subscribers[topic]
	if !ok {
		return
	}

	if s.mode == mailbox.Shared {

		for k := range s.consumers {
			if !s.consumers[k].IsExited() {
				s.consumers[k].PutMsg(msg)
			}
		}

	} else if s.mode == mailbox.Competition {
		s.consumers[rand.Intn(len(s.consumers))].PutMsg(msg)
	}
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
