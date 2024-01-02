package pubsubnsq

import (
	"fmt"
	"sync"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
)

const (
	// Name pub-sub plug-in name
	Name = "PubsubNsq"
)

type nsqPubsub struct {
	parm Parm

	sync.RWMutex

	info meta.ServiceInfo

	log *blog.Logger

	topicMap map[string]*pubsubTopic
}

func BuildWithOption(info meta.ServiceInfo, log *blog.Logger, opts ...Option) module.IPubsub {

	p := Parm{
		NsqLogLv:          nsq.LogLevelInfo,
		ConcurrentHandler: 1,
		ChannelLength:     2048,
	}

	for _, opt := range opts {
		opt(&p)
	}

	if len(p.NsqdAddress) != len(p.NsqdHttpAddress) {
		panic(fmt.Errorf("parm nsqd len(tcp addr) != len(http addr)"))
	}

	ps := &nsqPubsub{
		parm:     p,
		info:     info,
		log:      log,
		topicMap: make(map[string]*pubsubTopic),
	}

	return ps

}

func (nmb *nsqPubsub) GetTopic(name string) module.ITopic {
	var t *pubsubTopic

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

	//blog.Infof("Topic %v created", name)
	t.start()

	return t
}

func (nmb *nsqPubsub) Info() {

}

func (nmb *nsqPubsub) rmvTopic(name string) error {
	nmb.RLock()
	topic, ok := nmb.topicMap[name]
	nmb.RUnlock()

	if !ok {
		return fmt.Errorf("topic %v dose not exist", name)
	}

	//blog.Infof("deleting topic %v", name)
	err := topic.exit()
	if err != nil {
		return err
	}

	nmb.Lock()
	delete(nmb.topicMap, name)
	nmb.Unlock()

	return nil
}
