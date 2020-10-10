package balancer

import (
	"sync"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/mailbox"
)

// Group 负载均衡组管理器
type Group struct {
	m       sync.Map
	builder module.Builder
	mb      mailbox.IMailbox
}

var (
	bg *Group
)

// NewGroup 创建负载均衡组
func NewGroup(builder module.Builder, mb mailbox.IMailbox) *Group {

	bg = &Group{
		builder: builder,
		mb:      mb,
	}

	return bg
}

// Get 通过
func Get(nodName string) IBalancer {
	wb, ok := bg.m.Load(nodName)
	if !ok {
		wb, _ = bg.builder.Build(nodName, bg.mb)
		bg.m.Store(nodName, wb)
		log.Debugf("add balance group %s", nodName)
	}

	return wb.(IBalancer)
}
