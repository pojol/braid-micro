package module

import (
	"strings"

	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
)

// module type
const (
	TyDiscover      = "discover"
	TyBalancer      = "balancer"
	TyBalancerGroup = "balancer_group"
	TyElector       = "elector"
	TyLinkCache     = "link-cache"
	TyTracer        = "tracer"
	TyClient        = "client"
	TyServer        = "server"
)

// Builder builder
type Builder interface {
	// Build 模块的构建阶段
	// 这个阶段主要职责是：构建数据结构，配置各种参数行为等...
	Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (IModule, error)

	Name() string
	Type() string
	AddOption(opt interface{})
}

// IModule module
type IModule interface {
	// Init 模块的初始化阶段
	// 这个阶段主要职责是：部署和检测支撑本模块的运行依赖等...
	Init() error

	// Run 模块的运行期
	// 这个阶段主要职责是：主要用于提供周期性服务，一般会运行在goroutine中。
	Run()

	// Close 关闭模块
	// 这个阶段主要职责是：关闭本模块，并释放模块中依赖的各种资源。
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
