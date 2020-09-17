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
	"github.com/pojol/braid/plugin/balancerswrr"
	"github.com/pojol/braid/plugin/electorconsul"
	"github.com/pojol/braid/plugin/electork8s"
	"github.com/pojol/braid/plugin/grpcclient"
	"github.com/pojol/braid/plugin/grpcserver"
	"github.com/pojol/braid/plugin/linkerredis"
	"github.com/pojol/braid/plugin/pubsubnsq"
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

// BalancerBySwrr 基于平滑加权负载均衡
func BalancerBySwrr() Plugin {
	return func(b *Braid) {
		b.balancerBuilder = balancer.GetBuilder(balancerswrr.BalancerName)
	}
}

// LinkerByRedis 基于redis实现的链路缓存机制
func LinkerByRedis() Plugin {
	return func(b *Braid) {
		b.linkerBuilder = linkcache.GetBuilder(linkerredis.LinkerName)
		b.linkerBuilder.SetCfg(linkerredis.Config{
			ServiceName: b.cfg.Name,
		})
	}
}

// ElectorByConsul 基于consul实现的elector
func ElectorByConsul(consulAddr string) Plugin {
	return func(b *Braid) {

		b.electorBuild = elector.GetBuilder(electorconsul.ElectionName)
		if consulAddr == "" {
			consulAddr = "http://127.0.0.1:8500"
		}

		b.electorBuild.SetCfg(electorconsul.Cfg{
			Address:           consulAddr,
			Name:              b.cfg.Name,
			LockTick:          time.Second * 2,
			RefushSessionTick: time.Second * 5,
		})
	}
}

// ElectorByK8s 基于k8s实现的elector
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

// PubsubByNsq 构建pubsub
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
