// 实现文件 balancergroupbase 负载均衡管理器，主要用于统筹管理 服务:负载均衡算法
package balancergroupbase

import (
	"errors"
	"sync"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
)

const (
	// Name 基础的负载均衡容器实现
	Name = "BalancerGroupBase"
)

type baseBalancerGroupBuilder struct {
	opts []interface{}
}

func (b *baseBalancerGroupBuilder) Name() string {
	return Name
}

func (b *baseBalancerGroupBuilder) Type() string {
	return module.TyBalancerGroup
}

func newBaseBalancerGroup() module.Builder {
	return &baseBalancerGroupBuilder{}
}

func (b *baseBalancerGroupBuilder) AddOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *baseBalancerGroupBuilder) Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (module.IModule, error) {

	p := Parm{}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	bbg := &baseBalancerGroup{
		serviceName: serviceName,
		parm:        p,
		mb:          mb,
		logger:      logger,
		group:       make(map[string]*targetBalancerMap),
	}
	for _, strategy := range p.strategies {
		bbg.group[strategy] = &targetBalancerMap{
			targets: make(map[string]balancer.IBalancer),
		}
	}

	return bbg, nil
}

type targetBalancerMap struct {
	targets map[string]balancer.IBalancer
}

func (tbm *targetBalancerMap) get(target string) (balancer.IBalancer, error) {
	if _, ok := tbm.targets[target]; ok {
		return tbm.targets[target], nil
	}

	return nil, errors.New("can't find balancer, with target")
}

func (tbm *targetBalancerMap) exist(serviceName string) bool {
	if _, ok := tbm.targets[serviceName]; !ok {
		return false
	}

	return true
}

type baseBalancerGroup struct {
	mb          mailbox.IMailbox
	serviceName string

	parm Parm

	serviceUpdate mailbox.IChannel

	logger logger.ILogger

	// Strategy, Target
	group map[string]*targetBalancerMap

	lock sync.RWMutex
}

func (bbg *baseBalancerGroup) Init() error {

	bbg.serviceUpdate = bbg.mb.GetTopic(discover.ServiceUpdate).Sub(Name)

	return nil
}

func (bbg *baseBalancerGroup) Run() {

	bbg.serviceUpdate.Arrived(func(msg *mailbox.Message) {
		dmsg := discover.DecodeUpdateMsg(msg)
		if dmsg.Event == discover.EventAddService {
			bbg.lock.Lock()
			for strategy := range bbg.group {

				if !bbg.group[strategy].exist(dmsg.Nod.Name) {
					b := balancer.GetBuilder(strategy)
					ib, _ := b.Build(bbg.logger)
					bbg.group[strategy].targets[dmsg.Nod.Name] = ib
					bbg.logger.Debugf("add service %s by strategy %s", dmsg.Nod.Name, strategy)
				}

				bbg.group[strategy].targets[dmsg.Nod.Name].Add(dmsg.Nod)
			}
			bbg.lock.Unlock()
		} else if dmsg.Event == discover.EventRemoveService {
			bbg.lock.Lock()

			for k := range bbg.group {
				if _, ok := bbg.group[k]; ok {
					b, err := bbg.group[k].get(dmsg.Nod.Name)
					if err != nil {
						bbg.logger.Errorf("remove service err %s", err.Error())
						break
					}

					b.Rmv(dmsg.Nod)
				}
			}
			bbg.lock.Unlock()
		} else if dmsg.Event == discover.EventUpdateService {
			bbg.lock.Lock()

			for k := range bbg.group {
				if _, ok := bbg.group[k]; ok {
					b, err := bbg.group[k].get(dmsg.Nod.Name)
					if err != nil {
						bbg.logger.Errorf("update service err %s", err.Error())
						break
					}

					b.Update(dmsg.Nod)
				}
			}
			bbg.lock.Unlock()
		}
	})

}

func (bbg *baseBalancerGroup) Pick(ty string, target string) (discover.Node, error) {

	bbg.lock.RLock()
	defer bbg.lock.RUnlock()

	var nod discover.Node

	if _, ok := bbg.group[ty]; ok {

		b, err := bbg.group[ty].get(target)
		if err != nil {
			return nod, err
		}

		return b.Pick()
	}

	return nod, errors.New("can't find balancer, with strategy")
}

func (bbg *baseBalancerGroup) Close() {

}

func init() {
	module.Register(newBaseBalancerGroup())
}
