package pubsubredis

import (
	"sync"

	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
)

type redisPubsub struct {
	info meta.ServiceInfo
	parm Parm

	sync.RWMutex
	log    *blog.Logger
	client *redis.Client

	topicMap map[string]*redisTopic
}

func BuildWithOption(info meta.ServiceInfo, log *blog.Logger, cli *redis.Client, opts ...Option) module.IPubsub {

	p := Parm{}

	for _, opt := range opts {
		opt(&p)
	}

	ps := &redisPubsub{
		info:     info,
		client:   cli,
		log:      log,
		parm:     p,
		topicMap: make(map[string]*redisTopic),
	}

	return ps
}

func (nps *redisPubsub) Info() {

}

func (nps *redisPubsub) GetTopic(name string) module.ITopic {
	var t *redisTopic

	nps.RLock()
	t, ok := nps.topicMap[name]
	nps.RUnlock()
	if ok {
		return t
	}

	nps.Lock()
	t = newTopic(name, nps.client, nps, nps.log)
	nps.Unlock()

	return t
}
