// 实现文件 balancernormal 负载均衡管理器，主要用于统筹管理 服务:负载均衡算法
package balancer

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/pojol/braid-go/depend/pubsub"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/service"
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

func (s *balancerStrategy) Get(strategy string) (service.Node, error) {
	if strategy == StrategyRandom {
		return s.randomPicker.Get()
	} else if strategy == StrategySwrr {
		return s.swrrPicker.Get()
	}
	return service.Node{}, fmt.Errorf("not picker strategy %v", strategy)
}

func (s *balancerStrategy) Add(nod service.Node) {
	s.randomPicker.Add(nod)
	s.swrrPicker.Add(nod)
}

func (s *balancerStrategy) Rmv(nod service.Node) {
	s.randomPicker.Rmv(nod)
	s.swrrPicker.Rmv(nod)
}

func (s *balancerStrategy) Update(nod service.Node) {
	s.swrrPicker.Update(nod)
}

type baseBalancerGroup struct {
	ps          pubsub.IPubsub
	serviceName string

	serviceUpdate pubsub.IChannel

	picker map[string]*balancerStrategy

	sync.RWMutex
}

func BuildWithOption(name string, ps pubsub.IPubsub, opts ...Option) IBalancer {

	p := &Parm{}
	for _, opt := range opts {
		opt(p)
	}

	rand.Seed(time.Now().UnixNano())
	bbg := &baseBalancerGroup{
		serviceName: name,
		ps:          ps,
		picker:      make(map[string]*balancerStrategy),
	}

	return bbg
}

func (bbg *baseBalancerGroup) Init() {

	bbg.serviceUpdate = bbg.ps.GetTopic(service.TopicServiceUpdate).Sub(Name)

}

func (bbg *baseBalancerGroup) Run() {

	bbg.serviceUpdate.Arrived(func(msg *pubsub.Message) {
		dmsg := service.DiscoverDecodeUpdateMsg(msg)
		if dmsg.Event == discover.EventAddService {
			bbg.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; !ok {
				fmt.Println("add service", dmsg.Nod.Name)
				bbg.picker[dmsg.Nod.Name] = &balancerStrategy{
					randomPicker: &randomBalancer{},
					swrrPicker:   &swrrBalancer{},
				}
			}

			bbg.picker[dmsg.Nod.Name].Add(dmsg.Nod)

			bbg.Unlock()
		} else if dmsg.Event == discover.EventRemoveService {
			bbg.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; ok {
				bbg.picker[dmsg.Nod.Name].Rmv(dmsg.Nod)
			}

			bbg.Unlock()
		} else if dmsg.Event == discover.EventUpdateService {
			bbg.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; ok {
				bbg.picker[dmsg.Nod.Name].Update(dmsg.Nod)
			}

			bbg.Unlock()
		}
	})

}

func (bbg *baseBalancerGroup) Pick(strategy string, target string) (service.Node, error) {

	bbg.RLock()
	defer bbg.RUnlock()

	var nod service.Node

	if _, ok := bbg.picker[target]; ok {
		if strategy == StrategyRandom {
			return bbg.picker[target].randomPicker.Get()
		} else if strategy == StrategySwrr {
			return bbg.picker[target].swrrPicker.Get()
		}
	}

	return nod, errors.New("can't find balancer, with strategy")
}

func (bbg *baseBalancerGroup) Close() {

}
