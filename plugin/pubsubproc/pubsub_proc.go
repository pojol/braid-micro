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
		subs: make(map[string][]*braidsync.Unbounded),
	}

	return ps
}

func (*procPubsubBuilder) Name() string {
	return PubsubName
}

type procPubsub struct {
	sync.RWMutex
	subs map[string][]*braidsync.Unbounded
}

func (kps *procPubsub) Sub(topic string) *braidsync.Unbounded {
	kps.Lock()
	defer kps.Unlock()

	ch := braidsync.NewUnbounded()
	kps.subs[topic] = append(kps.subs[topic], ch)

	return ch
}

func (kps *procPubsub) Pub(topic string, msg interface{}) {
	kps.RLock()
	defer kps.RUnlock()

	for _, ch := range kps.subs[topic] {
		ch.Put(msg)
	}
}

func init() {
	pubsub.Register(newProcPubsub())
}
