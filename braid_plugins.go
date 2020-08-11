package braid

import (
	"time"

	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linker"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/plugin/balancer"
	"github.com/pojol/braid/plugin/balancer/swrrbalancer"
	"github.com/pojol/braid/plugin/discoverconsul"
	"github.com/pojol/braid/plugin/electorconsul"
	"github.com/pojol/braid/plugin/electork8s"
	"github.com/pojol/braid/plugin/linker/redislinker"
	"github.com/pojol/braid/plugin/rpc/grpcclient"
	"github.com/pojol/braid/plugin/rpc/grpcserver"
)

type config struct {
	Name string
}

// Plugin wraps
type Plugin func(*Braid)

// DiscoverByConsul 使用consul作为发现器支持
func DiscoverByConsul(address string, options ...discoverconsul.Option) Plugin {
	return func(b *Braid) {

		cfg := discoverconsul.Cfg{
			Name:     b.cfg.Name,
			Tag:      "braid",
			Interval: time.Second * 2,
			Address:  address,
		}

		for _, opt := range options {
			opt(&cfg)
		}

		b.discoverBuilder = discover.GetBuilder(discoverconsul.DiscoverName)
		err := b.discoverBuilder.SetCfg(cfg)
		if err != nil {
			// Fatal log
		}
	}
}

// BalancerBySwrr 基于平滑加权负载均衡
func BalancerBySwrr() Plugin {
	return func(b *Braid) {
		b.balancerBuilder = balancer.GetBuilder(swrrbalancer.BalancerName)
	}
}

// LinkerByRedis 基于redis实现的链路缓存机制
func LinkerByRedis() Plugin {
	return func(b *Braid) {
		b.linkerBuilder = linker.GetBuilder(redislinker.LinkerName)
	}
}

// ElectorByConsul 基于consul实现的elector
func ElectorByConsul() Plugin {
	return func(b *Braid) {
		b.electorBuild = elector.GetBuilder(electorconsul.ElectionName)
		b.electorBuild.SetCfg(electorconsul.Cfg{
			Address: "http://127.0.0.1:8500",
			Name:    b.cfg.Name,
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

// GRPCClient rpc-client
func GRPCClient(opts ...grpcclient.Option) Plugin {
	return func(b *Braid) {

		cfg := grpcclient.Config{
			PoolInitNum:  8,
			PoolCapacity: 32,
			PoolIdle:     120,
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
			Tracing:       false,
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
