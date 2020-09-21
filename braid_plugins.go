package braid

import (
	"time"

	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/pubsub"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
	"github.com/pojol/braid/plugin/grpcclient"
	"github.com/pojol/braid/plugin/grpcserver"
)

type config struct {
	Name string
}

// Plugin wraps
type Plugin func(*Braid)

// Discover plugin
func Discover(builderName string, opts ...interface{}) Plugin {
	return func(b *Braid) {
		b.discoverBuilder = discover.GetBuilder(builderName)

		for _, opt := range opts {
			b.discoverBuilder.AddOption(opt)
		}
	}
}

// Balancer plugin
func Balancer(builderName string, opts ...interface{}) Plugin {
	return func(b *Braid) {
		b.balancerBuilder = balancer.GetBuilder(builderName)
		for _, opt := range opts {
			b.balancerBuilder.AddOption(opt)
		}
	}
}

// LinkCache plugin
func LinkCache(builderName string, opts ...interface{}) Plugin {

	return func(b *Braid) {
		b.linkerBuilder = linkcache.GetBuilder(builderName)
		for _, opt := range opts {
			b.linkerBuilder.AddOption(opt)
		}
	}

}

// Elector plugin
func Elector(builderName string, opts ...interface{}) Plugin {
	return func(b *Braid) {
		b.electorBuild = elector.GetBuilder(builderName)
		for _, opt := range opts {
			b.electorBuild.AddOption(opt)
		}
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

// Pubsub plugin
func Pubsub(builderName string, opts ...interface{}) Plugin {

	return func(b *Braid) {
		b.pubsubBuilder = pubsub.GetBuilder(builderName)
		for _, opt := range opts {
			b.pubsubBuilder.AddOption(opt)
		}
	}

}

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
func GRPCClient(opts ...grpcclient.Option) Plugin {
	return func(b *Braid) {

		cfg := grpcclient.Config{
			PoolInitNum:  128,
			PoolCapacity: 1024,
			PoolIdle:     time.Second * 120,
		}

		for _, opt := range opts {
			opt(&cfg)
		}

		b.clientBuilder = client.GetBuilder(grpcclient.ClientName)
		b.clientBuilder.SetCfg(cfg)
	}
}

// GRPCServer rpc-server
func GRPCServer(opts ...grpcserver.Option) Plugin {
	return func(b *Braid) {
		cfg := grpcserver.Config{
			Name:          b.cfg.Name,
			ListenAddress: ":14222",
		}

		for _, opt := range opts {
			opt(&cfg)
		}

		b.serverBuilder = server.GetBuilder(grpcserver.ServerName)
		b.serverBuilder.SetCfg(cfg)
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
