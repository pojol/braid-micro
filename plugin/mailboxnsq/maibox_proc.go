package mailboxnsq

import (
	"sync"

	"github.com/google/uuid"
	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/internal/buffer"
	"github.com/pojol/braid/module/mailbox"
)

type procSubscriber struct {
	isShared    bool
	competition mailbox.IConsumer
	shared      sync.Map
}

type procConsumer struct {
	ID     string
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

func (ns *procSubscriber) AddCompetition() (mailbox.IConsumer, error) {

	if ns.competition == nil {
		ns.competition = &procConsumer{
			ID:     uuid.New().String(),
			buff:   buffer.NewUnbounded(),
			exitCh: braidsync.NewSwitch(),
		}
	}

	return ns.competition, nil
}

func (ns *procSubscriber) AddShared() (mailbox.IConsumer, error) {

	consumer := &procConsumer{
		ID:     uuid.New().String(),
		buff:   buffer.NewUnbounded(),
		exitCh: braidsync.NewSwitch(),
	}

	ns.isShared = true

	ns.shared.Store(consumer.ID, consumer)
	return consumer, nil
}

func (nmb *nsqMailbox) ProcPub(topic string, msg *mailbox.Message) {
	ss, ok := nmb.psubsrcibers.Load(topic)
	if !ok {
		return
	}

	sub, ok := ss.(*procSubscriber)
	if !ok {
		return
	}

	if sub.isShared {
		sub.shared.Range(func(key, value interface{}) bool {
			consumer, _ := value.(mailbox.IConsumer)
			if !consumer.IsExited() {
				consumer.PutMsg(msg)
			} else {
				nmb.psubsrcibers.Delete(topic)
			}

			return true
		})
	} else {
		if !sub.competition.IsExited() {
			sub.competition.PutMsg(msg)
		}
	}
}

func (nmb *nsqMailbox) ProcSub(topic string) mailbox.ISubscriber {

	ss, ok := nmb.psubsrcibers.Load(topic)
	if !ok {
		ss = &procSubscriber{}
		nmb.psubsrcibers.Store(topic, ss)
	}

	return ss.(mailbox.ISubscriber)

}
