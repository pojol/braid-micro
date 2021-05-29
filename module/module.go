package module

import (
	"strings"
)

type ModuleType int32

const (
	Discover ModuleType = iota + 1
	Balancer
	Elector
	Linkcache
	Tracer
	Client
	Server
	Pubsub
	Logger
)

// Builder builder
type IBuilder interface {
	Build(name string, buildOpts ...interface{}) interface{}

	Name() string
	Type() ModuleType

	AddModuleOption(opt interface{})
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
	m = make(map[string]IBuilder)
)

// Register 注册balancer
func Register(b IBuilder) {
	m[strings.ToLower(b.Name())] = b
}

// GetBuilder 获取构建器
func GetBuilder(name string) IBuilder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}
