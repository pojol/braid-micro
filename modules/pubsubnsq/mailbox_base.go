// 实现文件 基于 nsq 实现的 pubsub
package pubsubnsq

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/moduleparm"
)

const (
	// Name pub-sub plug-in name
	Name = "PubsubNsq"
)

type nsqPubsubBuilder struct {
	opts []interface{}
}

func newNsqPubsub() module.IBuilder {
	return &nsqPubsubBuilder{}
}

func (nb *nsqPubsubBuilder) AddModuleOption(opt interface{}) {
	nb.opts = append(nb.opts, opt)
}

func (nb *nsqPubsubBuilder) Name() string {
	return Name
}

func (nb *nsqPubsubBuilder) Type() module.ModuleType {
	return module.Pubsub
}

func (nb *nsqPubsubBuilder) Build(name string, buildOpts ...interface{}) interface{} {
	bp := moduleparm.BuildParm{}
	for _, opt := range buildOpts {
		opt.(moduleparm.Option)(&bp)
	}

	p := Parm{
		ServiceName:       name,
		nsqLogLv:          nsq.LogLevelWarning,
		ConcurrentHandler: 1,
	}
	for _, opt := range nb.opts {
		opt.(Option)(&p)
	}

	rand.Seed(time.Now().UnixNano())
	if len(p.NsqdAddress) != len(p.NsqdHttpAddress) {
		panic(fmt.Errorf("parm nsqd len(tcp addr) != len(http addr)"))
	}

	nsqm := &nsqPubsub{
		parm:     p,
		log:      bp.Logger,
		topicMap: make(map[string]*pubsubTopic),
	}

	return nsqm
}

type nsqPubsub struct {
	parm Parm
	log  logger.ILogger

	sync.RWMutex

	topicMap map[string]*pubsubTopic
}

func (nmb *nsqPubsub) RegistTopic(name string, scope pubsub.ScopeTy) (pubsub.ITopic, error) {

	nmb.Lock()
	t, ok := nmb.topicMap[name]
	if ok {
		nmb.Unlock()
		return t, nil
	}

	t = newTopic(name, scope, nmb)
	nmb.topicMap[name] = t
	nmb.Unlock()
	nmb.log.Infof("Topic %v created", name)

	// start loop
	t.start()

	return t, nil
}

func (nmb *nsqPubsub) GetTopic(name string) pubsub.ITopic {

	nmb.RLock()
	t, ok := nmb.topicMap[name]
	nmb.RUnlock()
	if ok {
		return t
	}

	nt, err := nmb.RegistTopic(name, pubsub.ScopeProc)
	if err != nil {
		panic(err)
	}
	nmb.log.Warnf("Get topic warning %v undefined! register proc topic", name)

	return nt
}

func (nmb *nsqPubsub) RemoveTopic(name string) error {
	nmb.RLock()
	topic, ok := nmb.topicMap[name]
	nmb.RUnlock()

	if !ok {
		return fmt.Errorf("topic %v dose not exist", name)
	}

	nmb.log.Infof("deleting topic %v", name)
	topic.Exit()

	nmb.Lock()
	delete(nmb.topicMap, name)
	nmb.Unlock()

	return nil
}

func init() {
	module.Register(newNsqPubsub())
}
