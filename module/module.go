package module

import (
	"strings"

	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
)

// module type
const (
	TyDiscover      = "discover"
	TyBalancer      = "balancer"
	TyBalancerGroup = "balancer_group"
	TyElector       = "elector"
	TyLinkCache     = "link-cache"
	TyTracer        = "tracer"
)

// Builder builder
type Builder interface {
	Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (IModule, error)
	Name() string
	Type() string
	AddOption(opt interface{})
}

// IModule module
type IModule interface {
	Init()
	Run()
	Close()
}

var (
	m = make(map[string]Builder)
)

// Register 注册balancer
func Register(b Builder) {
	m[strings.ToLower(b.Name())] = b
}

// GetBuilder 获取构建器
func GetBuilder(name string) Builder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}
