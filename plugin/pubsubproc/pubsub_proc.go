package pubsubproc

import (
	"sync"

	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/module/pubsub"
)

const (
	// PubsubName 进程内的消息通知
	PubsubName = "ProcPubsub"
)

type procPubsubBuilder struct {
}

func newProcPubsub() pubsub.Builder {
	return &procPubsubBuilder{}
}

func (*procPubsubBuilder) Build() pubsub.IPubsub {

	ps := &procPubsub{
		consumer: make(map[string][]pubsub.IConsumer),
	}

	return ps
}

func (*procPubsubBuilder) Name() string {
	return PubsubName
}

type procPubsub struct {
	sync.RWMutex
	consumer map[string][]pubsub.IConsumer
}

// Consumer 消费者
type procConsumer struct {
	buff   *braidsync.Unbounded
	exitCh *braidsync.Switch
}

func (c *procConsumer) AddHandler(handler pubsub.HandlerFunc) {

	go func() {
		for {
			select {
			case msg := <-c.buff.Get():
				handler(msg.(*pubsub.Message))
				c.buff.Load()
			case <-c.exitCh.Done():
			}

			if c.exitCh.HasOpend() {
				return
			}
		}
	}()

}

func (c *procConsumer) Exit() {
	c.exitCh.Open()
}

func (c *procConsumer) IsExited() bool {
	return c.exitCh.HasOpend()
}

func (c *procConsumer) PutMsg(msg *pubsub.Message) {
	c.buff.Put(msg)
}

func (kps *procPubsub) Sub(topic string) pubsub.IConsumer {
	kps.Lock()
	defer kps.Unlock()

	c := &procConsumer{
		buff:   braidsync.NewUnbounded(),
		exitCh: braidsync.NewSwitch(),
	}
	kps.consumer[topic] = append(kps.consumer[topic], c)

	return c
}

func (kps *procPubsub) Pub(topic string, msg *pubsub.Message) {
	kps.RLock()
	defer kps.RUnlock()

	for i := 0; i < len(kps.consumer[topic]); i++ {
		if kps.consumer[topic][i].IsExited() {
			kps.consumer[topic] = append(kps.consumer[topic][:i], kps.consumer[topic][i+1:]...)
			i--
		} else {
			kps.consumer[topic][i].PutMsg(msg)
		}
	}

}

func init() {
	pubsub.Register(newProcPubsub())
}
