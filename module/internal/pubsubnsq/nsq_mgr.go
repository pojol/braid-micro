package pubsubnsq

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/module/pubsub"
)

const (
	// Name pub-sub plug-in name
	Name = "PubsubNsq"
)

func BuildWithOption(name string, log *blog.Logger, opts ...pubsub.NsqOption) pubsub.IPubsub {

	p := pubsub.NsqParm{
		ServiceName:       name,
		NsqLogLv:          nsq.LogLevelWarning,
		ConcurrentHandler: 1,
		ChannelLength:     2048,
	}
	for _, opt := range opts {
		opt(&p)
	}

	rand.Seed(time.Now().UnixNano())
	if len(p.NsqdAddress) != len(p.NsqdHttpAddress) {
		panic(fmt.Errorf("parm nsqd len(tcp addr) != len(http addr)"))
	}

	nsqm := &nsqPubsub{
		parm:     p,
		log:      log,
		topicMap: make(map[string]*pubsubTopic),
	}

	return nsqm
}

type nsqPubsub struct {
	parm pubsub.NsqParm

	sync.RWMutex

	log *blog.Logger

	topicMap map[string]*pubsubTopic
}

func (nmb *nsqPubsub) getTopic(name string) pubsub.ITopic {

	var t *pubsubTopic

	nmb.Lock()
	t = newTopic(name, nmb)
	nmb.topicMap[name] = t
	nmb.Unlock()

	//blog.Infof("Topic %v created", name)
	t.start()

	return t
}

func (nmb *nsqPubsub) Topic(name string) pubsub.ITopic {
	return nmb.getTopic(name)
}

func (nmb *nsqPubsub) rmvTopic(name string) error {
	nmb.RLock()
	topic, ok := nmb.topicMap[name]
	nmb.RUnlock()

	if !ok {
		return fmt.Errorf("topic %v dose not exist", name)
	}

	//blog.Infof("deleting topic %v", name)
	err := topic.Exit()
	if err != nil {
		return err
	}

	nmb.Lock()
	delete(nmb.topicMap, name)
	nmb.Unlock()

	return nil
}
