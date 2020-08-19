package pubsubnsq

import (
	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid/module/pubsub"
)

const (
	// PubsubName 进程内的消息通知
	PubsubName = "NsqPubsub"
)

type procPubsubBuilder struct {
}

func newNsqPubsub() pubsub.Builder {
	return &procPubsubBuilder{}
}

func (*procPubsubBuilder) Build() pubsub.IPubsub {

	producer, err := nsq.NewProducer("192.168.50.201:4150", nsq.NewConfig())
	if err != nil {
		panic(err)
	}

	ps := &nsqPubsub{
		producer: producer,
	}

	return ps
}

func (*procPubsubBuilder) Name() string {
	return PubsubName
}

type nsqPubsub struct {
	producer *nsq.Producer
}

// Consumer 消费者
type nsqConsumer struct {
	consumer *nsq.Consumer
}

type nsqSubscriber struct {
}

func (ns *nsqSubscriber) AddConsumer() pubsub.IConsumer {

	return nil
}

func (ns *nsqSubscriber) AppendConsumer() pubsub.IConsumer {
	return nil
}

func (ns *nsqSubscriber) PutMsg(msg *pubsub.Message) {

}

func (c *nsqConsumer) OnArrived(handler pubsub.HandlerFunc) {

}

func (c *nsqConsumer) Exit() {
}

func (c *nsqConsumer) IsExited() bool {
	return false
}

func (c *nsqConsumer) PutMsg(msg *pubsub.Message) {
}

func (kps *nsqPubsub) Sub(topic string) pubsub.ISubscriber {
	s := &nsqSubscriber{}
	return s
}

func (kps *nsqPubsub) Pub(topic string, msg *pubsub.Message) {

}

func init() {
	pubsub.Register(newNsqPubsub())
}
