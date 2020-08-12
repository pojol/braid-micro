package balancer

import (
	"sync"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/pubsub"
)

// Group 负载均衡组管理器
type Group struct {
	m       sync.Map
	builder Builder
	ps      pubsub.IPubsub
}

var (
	bg *Group
)

// NewGroup 创建负载均衡组
func NewGroup(builder Builder, pubsub pubsub.IPubsub) *Group {

	bg = &Group{
		builder: builder,
		ps:      pubsub,
	}

	return bg
}

// Get 通过
func Get(nodName string) Balancer {
	wb, ok := bg.m.Load(nodName)
	if !ok {
		wb = bg.builder.Build(bg.ps)
		bg.m.Store(nodName, wb)
		log.Debugf("add balance group %s", nodName)
	}

	return wb.(Balancer)
}
