package pubsub

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/depend/blog"
)

const (
	// Name pub-sub plug-in name
	Name = "PubsubNsq"
)

func BuildWithOption(name string, opts ...Option) IPubsub {

	p := Parm{
		ServiceName:       name,
		nsqLogLv:          nsq.LogLevelWarning,
		ConcurrentHandler: 1,
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
		topicMap: make(map[string]*pubsubTopic),
	}

	return nsqm
}

type nsqPubsub struct {
	parm Parm

	sync.RWMutex

	topicMap map[string]*pubsubTopic
}

func (nmb *nsqPubsub) GetTopic(name string) ITopic {

	nmb.RLock()
	t, ok := nmb.topicMap[name]
	nmb.RUnlock()
	if ok {
		return t
	}

	nmb.Lock()
	t = newTopic(name, nmb)
	nmb.topicMap[name] = t
	nmb.Unlock()

	blog.Infof("Topic %v created", name)

	// start loop
	t.start()

	return t
}

func (nmb *nsqPubsub) RemoveTopic(name string) error {
	nmb.RLock()
	topic, ok := nmb.topicMap[name]
	nmb.RUnlock()

	if !ok {
		return fmt.Errorf("topic %v dose not exist", name)
	}

	blog.Infof("deleting topic %v", name)
	topic.Exit()

	nmb.Lock()
	delete(nmb.topicMap, name)
	nmb.Unlock()

	return nil
}
