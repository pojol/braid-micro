package braid

import (
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
)

type config struct {
	Name string
}

// Plugin wraps
type Plugin func(*Braid)

// Discover plugin
func Discover(builderName string, opts ...interface{}) Plugin {
	return func(b *Braid) {

		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		b.builders = append(b.builders, builder)
	}
}

// Balancer plugin
func Balancer(builderName string, opts ...interface{}) Plugin {
	return func(b *Braid) {
		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		balancer.NewGroup(builder, b.mailbox, b.logger)
		b.builders = append(b.builders, builder)
	}
}

// LinkCache plugin
func LinkCache(builderName string, opts ...interface{}) Plugin {

	return func(b *Braid) {
		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		b.builders = append(b.builders, builder)
	}

}

// Elector plugin
func Elector(builderName string, opts ...interface{}) Plugin {
	return func(b *Braid) {

		builder := module.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}
		b.builders = append(b.builders, builder)
	}
}

// ElectorByK8s 基于k8s实现的elector
/*
func ElectorByK8s(kubeconfig string, nodid string) Plugin {
	return func(b *Braid) {
		b.electorBuild = elector.GetBuilder(electork8s.ElectionName)
		b.electorBuild.SetCfg(electork8s.Cfg{
			KubeCfg:     kubeconfig,
			NodID:       nodid,
			Namespace:   "default",
			RetryPeriod: time.Second * 2,
		})
	}
}
*/

// PubsubByNsq 构建pubsub
/*
func PubsubByNsq(lookupAddres []string, addr []string, opts ...pubsubnsq.Option) Plugin {
	return func(b *Braid) {
		b.pubsubBuilder = pubsub.GetBuilder(pubsubnsq.PubsubName)
		cfg := pubsubnsq.NsqConfig{
			LookupAddres: lookupAddres,
			Addres:       addr,
		}

		for _, opt := range opts {
			opt(&cfg)
		}

		b.pubsubBuilder.SetCfg(cfg)
	}
}
*/

// GRPCClient rpc-client
func GRPCClient(builderName string, opts ...interface{}) Plugin {
	return func(b *Braid) {

		builder := client.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}

		b.clientBuilder = builder
	}
}

// GRPCServer rpc-server
func GRPCServer(builderName string, opts ...interface{}) Plugin {
	return func(b *Braid) {

		builder := server.GetBuilder(builderName)
		for _, opt := range opts {
			builder.AddOption(opt)
		}

		b.serverBuilder = builder
	}
}

// JaegerTracing jt
func JaegerTracing(protoOpt tracer.Option, opts ...tracer.Option) Plugin {
	return func(b *Braid) {

		t, err := tracer.New(b.cfg.Name, protoOpt, opts...)
		if err != nil {
			panic(err)
		}

		b.tracer = t
	}
}
