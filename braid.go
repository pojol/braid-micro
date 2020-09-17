package braid

import (
	"fmt"

	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/pubsub"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
	"github.com/pojol/braid/plugin/discoverconsul"
	"github.com/pojol/braid/plugin/pubsubproc"
)

// Braid framework instance
type Braid struct {
	cfg config

	clientBuilder client.Builder
	client        client.IClient

	serverBuilder server.Builder
	server        server.ISserver

	discoverBuilder discover.Builder
	discover        discover.IDiscover

	linkerBuilder linkcache.Builder
	linker        linkcache.ILinkCache

	electorBuild elector.Builder
	elector      elector.IElection

	balancerBuilder balancer.Builder

	pubsubBuilder pubsub.Builder
	pubsub        pubsub.IPubsub

	tracer *tracer.Tracer
}

var (
	braidGlobal *Braid
)

// New 构建braid
func New(name string) *Braid {

	braidGlobal = &Braid{
		cfg: config{
			Name: name,
		},
	}
	return braidGlobal
}

// RegistPlugin 注册插件
func (b *Braid) RegistPlugin(plugins ...Plugin) error {

	// install default
	var err error

	//
	for _, plugin := range plugins {
		plugin(braidGlobal)
	}

	pb, _ := pubsub.GetBuilder(pubsubproc.PubsubName).Build()

	// build

	if b.balancerBuilder != nil {
		balancer.NewGroup(b.balancerBuilder, pb)
	}

	if b.electorBuild != nil {
		b.elector, _ = b.electorBuild.Build()
	}

	if b.pubsubBuilder != nil {
		b.pubsub, _ = b.pubsubBuilder.Build()
	}

	if b.linkerBuilder != nil {
		if b.electorBuild == nil {
			fmt.Println("linker need depend elector")
		}
		if b.pubsubBuilder == nil {
			fmt.Println("linker need depend pubsub")
		}

		b.linker = b.linkerBuilder.Build(b.elector, b.pubsub)
	}

	if b.discoverBuilder != nil {
		if b.balancerBuilder == nil {
			fmt.Println("discover need depend balancer")
		}

		b.discoverBuilder.AddOption(discoverconsul.WithProcPubsub(pb))

		if b.linker != nil {
			b.discoverBuilder.AddOption(discoverconsul.WithLinkCache(b.linker))
		}

		if b.pubsub != nil {
			b.discoverBuilder.AddOption(discoverconsul.WithClusterPubsub(b.pubsub))
		}

		b.discover, err = b.discoverBuilder.Build(b.cfg.Name)
		if err != nil {
			panic(err)
		}
	}

	if b.serverBuilder != nil {
		b.server = b.serverBuilder.Build(b.tracer != nil)
	}

	if b.clientBuilder != nil {
		// check discover
		if b.discoverBuilder == nil {
			fmt.Println("rpc-client need depend discover")
		}

		b.client = b.clientBuilder.Build(b.linker, b.tracer != nil)
	}

	return nil
}

// Run 运行braid
func (b *Braid) Run() {

	if b.discover != nil {
		b.discover.Discover()
	}

	if b.elector != nil {
		b.elector.Run()
	}

	if b.server != nil {
		b.server.Run()
	}

}

// Client grpc-client
func Client() client.IClient {
	return braidGlobal.client
}

// Server grpc-server
func Server() server.ISserver {
	return braidGlobal.server
}

// Pubsub pub-sub
func Pubsub() pubsub.IPubsub {
	return braidGlobal.pubsub
}

// Close 关闭braid
func (b *Braid) Close() {

	if b.discover != nil {
		b.discover.Close()
	}

	if b.elector != nil {
		b.elector.Close()
	}

	if b.server != nil {
		b.server.Close()
	}

	if b.tracer != nil {
		b.tracer.Close()
	}

}
