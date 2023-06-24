// 实现文件 balancernormal 负载均衡管理器，主要用于统筹管理 服务:负载均衡算法
package balancer

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
)

const (
	// Name 基础的负载均衡容器实现
	Name = "BalancerNormal"

	StrategyRandom = "strategy_random"
	StrategySwrr   = "strategy_swrr"
)

type balancerStrategy struct {
	randomPicker IPicker
	swrrPicker   IPicker
}

func (s *balancerStrategy) Get(strategy string) (meta.Node, error) {
	if strategy == StrategyRandom {
		return s.randomPicker.Get()
	} else if strategy == StrategySwrr {
		return s.swrrPicker.Get()
	}
	return meta.Node{}, fmt.Errorf("not picker strategy %v", strategy)
}

func (s *balancerStrategy) Add(nod meta.Node) {
	s.randomPicker.Add(nod)
	s.swrrPicker.Add(nod)
}

func (s *balancerStrategy) Rmv(nod meta.Node) {
	s.randomPicker.Rmv(nod)
	s.swrrPicker.Rmv(nod)
}

func (s *balancerStrategy) Update(nod meta.Node) {
	s.swrrPicker.Update(nod)
}

type baseBalancerGroup struct {
	ps          module.IPubsub
	log         *blog.Logger
	serviceInfo meta.ServiceInfo

	serviceUpdate module.IChannel

	picker map[string]*balancerStrategy

	sync.RWMutex
}

func BuildWithOption(info meta.ServiceInfo, log *blog.Logger, ps module.IPubsub, opts ...Option) IBalancer {

	p := &Parm{}
	for _, opt := range opts {
		opt(p)
	}

	rand.Seed(time.Now().UnixNano())
	bbg := &baseBalancerGroup{
		serviceInfo: info,
		ps:          ps,
		log:         log,
		picker:      make(map[string]*balancerStrategy),
	}

	return bbg
}

func (bbg *baseBalancerGroup) Init() {

	bbg.serviceUpdate, _ = bbg.ps.GetTopic(meta.TopicDiscoverServiceUpdate).
		Sub(context.TODO(), meta.ModuleBalancer+"-"+bbg.serviceInfo.ID)

}

func (bbg *baseBalancerGroup) Run() {

	bbg.serviceUpdate.Arrived(func(msg *meta.Message) error {
		dmsg := meta.DecodeUpdateMsg(msg)
		if dmsg.Event == meta.TopicDiscoverServiceNodeAdd {
			bbg.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; !ok {
				bbg.picker[dmsg.Nod.Name] = &balancerStrategy{
					randomPicker: &randomBalancer{},
					swrrPicker:   &swrrBalancer{},
				}
			}

			bbg.picker[dmsg.Nod.Name].Add(dmsg.Nod)
			bbg.Unlock()
		} else if dmsg.Event == meta.TopicDiscoverServiceNodeRmv {
			bbg.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; ok {
				bbg.picker[dmsg.Nod.Name].Rmv(dmsg.Nod)
			}

			bbg.Unlock()
		} else if dmsg.Event == meta.TopicDiscoverServiceUpdate {
			bbg.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; ok {
				bbg.picker[dmsg.Nod.Name].Update(dmsg.Nod)
			}

			bbg.Unlock()
		}

		return nil
	})

}

func (bbg *baseBalancerGroup) Pick(strategy string, target string) (meta.Node, error) {

	bbg.RLock()
	defer bbg.RUnlock()

	var nod meta.Node
	var err error

	if _, ok := bbg.picker[target]; ok {
		if strategy == StrategyRandom {
			nod, err = bbg.picker[target].randomPicker.Get()
		} else if strategy == StrategySwrr {
			nod, err = bbg.picker[target].swrrPicker.Get()
		}
	}

	bbg.log.Infof("pick %s %s %v %v", strategy, target, nod, err)
	return nod, err
}

func (bbg *baseBalancerGroup) Close() {
	bbg.serviceUpdate.Close()
}
