package pubsubkafka

import (
	"sync"

	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/module/pubsub"
)

const (
	// PubsubName 选举器名称
	PubsubName = "KafkaPubsub"
)

type kafkaPubsubBuilder struct {
}

func newKafkaPubsub() pubsub.Builder {
	return &kafkaPubsubBuilder{}
}

func (*kafkaPubsubBuilder) Build() pubsub.IPubsub {

	ps := &kafkaPubsub{
		subs: make(map[string][]*braidsync.Unbounded),
	}

	return ps
}

func (*kafkaPubsubBuilder) Name() string {
	return PubsubName
}

type kafkaPubsub struct {
	sync.RWMutex
	subs map[string][]*braidsync.Unbounded
}

func (kps *kafkaPubsub) Sub(topic string) *braidsync.Unbounded {
	kps.Lock()
	defer kps.Unlock()

	ch := braidsync.NewUnbounded()
	kps.subs[topic] = append(kps.subs[topic], ch)

	return ch
}

func (kps *kafkaPubsub) Pub(topic string, msg interface{}) {
	kps.RLock()
	defer kps.RUnlock()

	for _, ch := range kps.subs[topic] {
		ch.Put(msg)
	}
}

func init() {
	pubsub.Register(newKafkaPubsub())
}
