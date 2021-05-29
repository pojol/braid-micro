// 实现文件 balancernormal 负载均衡管理器，主要用于统筹管理 服务:负载均衡算法
package balancernormal

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/moduleparm"
)

const (
	// Name 基础的负载均衡容器实现
	Name = "BalancerNormal"

	StrategyRandom = "strategy_random"
	StrategySwrr   = "strategy_swrr"
)

type baseBalanceBuilder struct {
	opts []interface{}
}

func (b *baseBalanceBuilder) Name() string {
	return Name
}

func (b *baseBalanceBuilder) Type() module.ModuleType {
	return module.Balancer
}

func newBaseBalancerGroup() module.IBuilder {
	return &baseBalanceBuilder{}
}

func (b *baseBalanceBuilder) AddModuleOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *baseBalanceBuilder) Build(name string, buildOpts ...interface{}) interface{} {

	bp := moduleparm.BuildParm{}
	for _, opt := range buildOpts {
		opt.(moduleparm.Option)(&bp)
	}

	p := Parm{}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	rand.Seed(time.Now().UnixNano())
	bbg := &baseBalancerGroup{
		serviceName: name,
		parm:        p,
		ps:          bp.PS,
		logger:      bp.Logger,
		picker:      make(map[string]*balancerStrategy),
	}

	return bbg
}

type balancerStrategy struct {
	randomPicker balancer.IPicker
	swrrPicker   balancer.IPicker
}

func (s *balancerStrategy) Get(strategy string) (discover.Node, error) {
	if strategy == StrategyRandom {
		return s.randomPicker.Get()
	} else if strategy == StrategySwrr {
		return s.swrrPicker.Get()
	}
	return discover.Node{}, fmt.Errorf("not picker strategy %v", strategy)
}

func (s *balancerStrategy) Add(nod discover.Node) {
	s.randomPicker.Add(nod)
	s.swrrPicker.Add(nod)
}

func (s *balancerStrategy) Rmv(nod discover.Node) {
	s.randomPicker.Rmv(nod)
	s.swrrPicker.Rmv(nod)
}

func (s *balancerStrategy) Update(nod discover.Node) {
	s.swrrPicker.Update(nod)
}

type baseBalancerGroup struct {
	ps          pubsub.IPubsub
	serviceName string

	parm Parm

	serviceUpdate pubsub.IChannel

	logger logger.ILogger

	picker map[string]*balancerStrategy

	lock sync.RWMutex
}

func (bbg *baseBalancerGroup) Init() error {

	bbg.serviceUpdate = bbg.ps.GetTopic(discover.ServiceUpdate).Sub(Name)

	return nil
}

func (bbg *baseBalancerGroup) Run() {

	bbg.serviceUpdate.Arrived(func(msg *pubsub.Message) {
		dmsg := discover.DecodeUpdateMsg(msg)
		if dmsg.Event == discover.EventAddService {
			bbg.lock.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; !ok {
				fmt.Println("add service", dmsg.Nod.Name)
				bbg.picker[dmsg.Nod.Name] = &balancerStrategy{
					randomPicker: &randomBalancer{logger: bbg.logger},
					swrrPicker:   &swrrBalancer{logger: bbg.logger},
				}
			}

			bbg.picker[dmsg.Nod.Name].Add(dmsg.Nod)

			bbg.lock.Unlock()
		} else if dmsg.Event == discover.EventRemoveService {
			bbg.lock.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; ok {
				bbg.picker[dmsg.Nod.Name].Rmv(dmsg.Nod)
			}

			bbg.lock.Unlock()
		} else if dmsg.Event == discover.EventUpdateService {
			bbg.lock.Lock()

			if _, ok := bbg.picker[dmsg.Nod.Name]; ok {
				bbg.picker[dmsg.Nod.Name].Update(dmsg.Nod)
			}

			bbg.lock.Unlock()
		}
	})

}

func (bbg *baseBalancerGroup) Pick(strategy string, target string) (discover.Node, error) {

	bbg.lock.RLock()
	defer bbg.lock.RUnlock()

	var nod discover.Node

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

func init() {
	module.Register(newBaseBalancerGroup())
}
