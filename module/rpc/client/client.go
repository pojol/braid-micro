// rpc-client 模块
//
package client

import (
	"context"
	"strings"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
)

// Builder grpc-client builder
type Builder interface {
	Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (IClient, error)
	Name() string
	AddOption(opt interface{})
}

// IClient rpc-client interface
type IClient interface {
	module.IModule

	// Invoke 发起一次 rpc 调用
	//
	// ctx 上下文，用于保存链路过程中的信息（主要用于分布式追踪
	//
	// target 目标服务的名称，用于获取定位到该服务的相关信息
	//
	// methon 目标服务中的方法，用于派送到目标服务的该方法中
	//
	// token 用户唯一凭据
	// 如果传入是空的值，则在路由到目标服务器时采用无状态的负载均衡方案（如随机。
	// 如果传入是用户的唯一凭据，则在路由的过程中采用有状态的负载均衡方案（默认提供的是平滑加权算法。
	// 如果在 braid 中注册了 linkcache 模块则通过 token 能保证此 token 在链路过程中选取的目标服务器是固定的。
	//
	// args 调用发送的参数
	//
	// reply 调用返回的参数
	//
	// opts 调用的可选项
	Invoke(
		ctx context.Context, target, methon, token string,
		args, reply interface{},
		opts ...interface{}) error
}

var (
	m = make(map[string]Builder)
)

// Register 注册 builder
func Register(b Builder) {
	m[strings.ToLower(b.Name())] = b
}

// GetBuilder 获取 builder
func GetBuilder(name string) Builder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}
