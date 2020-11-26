package braid

import (
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
)

type config struct {
	Name string
}

// Module wraps
type Module func(*Braid)

// Discover 服务发现模块，提供以下异步消息
// 	discover.AddService 发现新的服务
// 	discover.RmvService 某个服务下线
// 	discover.UpdateService 更新某个服务的权重信息
func Discover(builderName string, opts ...interface{}) Module {
	return func(b *Braid) {

		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		b.builders = append(b.builders, builder)
	}
}

// LinkCache 链路缓存模块，提供以下异步消息
//	linkcache.ServiceLinkNum 获得服务当前的链路数量
func LinkCache(builderName string, opts ...interface{}) Module {

	return func(b *Braid) {
		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		b.builders = append(b.builders, builder)
	}

}

// Elector 选举模块
// 	elector.StateChange 状态改变消息 Wait, Slave, Master
func Elector(builderName string, opts ...interface{}) Module {
	return func(b *Braid) {

		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		b.builders = append(b.builders, builder)
	}
}

// GRPCClient rpc-client
func GRPCClient(builderName string, opts ...interface{}) Module {
	return func(b *Braid) {

		builder := client.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}

		b.clientBuilder = builder
	}
}

// GRPCServer rpc-server
func GRPCServer(builderName string, opts ...interface{}) Module {
	return func(b *Braid) {

		builder := server.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}

		b.serverBuilder = builder
	}
}

// Tracing 分布式追踪模块
func Tracing(builderName string, opts ...interface{}) Module {
	return func(b *Braid) {
		builder := tracer.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}

		b.tracerBuilder = builder
	}
}
