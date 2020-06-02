package balancer

import (
	"sync"

	"github.com/pojol/braid/log"
)

// Group 负载均衡组管理器
type Group struct {
	m sync.Map

	builderName string
}

// NewGroup 创建负载均衡组
func NewGroup(opts ...Option) *Group {

	g := &Group{
		builderName: "WeightedRoundrobin",
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// Option group config wraps
type Option func(*Group)

// WithBuilder 通过其他balancer构建
func WithBuilder(builderName string) Option {
	return func(g *Group) {
		g.builderName = builderName
	}
}

// Get 通过
func (g *Group) Get(nodName string) Balancer {
	wb, ok := g.m.Load(nodName)
	if !ok {
		wb = newBalancerWrapper(GetBuilder(g.builderName))
		g.m.Store(nodName, wb)

		log.Debugf("add balance group %s", nodName)
	}

	return wb.(Balancer)
}
