package balancer

import (
	"sync"

	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
)

// Group 负载均衡组管理器
type Group struct {
	m       sync.Map
	builder module.Builder
	mb      mailbox.IMailbox
	logger  logger.ILogger
}

var (
	bg *Group
)

// NewGroup 创建负载均衡组
func NewGroup(builder module.Builder, mb mailbox.IMailbox, logger logger.ILogger) *Group {

	bg = &Group{
		builder: builder,
		mb:      mb,
		logger:  logger,
	}

	return bg
}

// Get 通过
func Get(nodName string) IBalancer {
	wb, ok := bg.m.Load(nodName)
	if !ok {
		wb, _ = bg.builder.Build(nodName, bg.mb, bg.logger)
		bg.m.Store(nodName, wb)
	}

	return wb.(IBalancer)
}
