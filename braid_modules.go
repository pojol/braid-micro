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

// Discover plugin
func Discover(builderName string, opts ...interface{}) Module {
	return func(b *Braid) {

		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		b.builders = append(b.builders, builder)
	}
}

// LinkCache plugin
func LinkCache(builderName string, opts ...interface{}) Module {

	return func(b *Braid) {
		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		b.builders = append(b.builders, builder)
	}

}

// Elector plugin
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

// JaegerTracing jt
func JaegerTracing(protoOpt tracer.Option, opts ...tracer.Option) Module {
	return func(b *Braid) {

		t, err := tracer.New(b.cfg.Name, protoOpt, opts...)
		if err != nil {
			panic(err)
		}

		b.tracer = t
	}
}
