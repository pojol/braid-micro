package braid

import (
	"fmt"

	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linker"
	"github.com/pojol/braid/module/pubsub"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
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

	linkerBuilder linker.Builder
	linker        linker.ILinker

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

	//
	for _, plugin := range plugins {
		plugin(braidGlobal)
	}

	b.pubsubBuilder = pubsub.GetBuilder(pubsubproc.PubsubName)
	b.pubsub = b.pubsubBuilder.Build()

	// build
	if b.discoverBuilder != nil {
		b.discover = b.discoverBuilder.Build(b.pubsub)
	}

	if b.balancerBuilder != nil {
		balancer.NewGroup(b.balancerBuilder, b.pubsub)
	}

	if b.electorBuild != nil {
		b.elector, _ = b.electorBuild.Build()
	}

	if b.serverBuilder != nil {
		b.server = b.serverBuilder.Build()
	}

	if b.linker != nil {
		if b.electorBuild != nil {

		}

	}

	if b.clientBuilder != nil {

		// check balancer
		if b.balancerBuilder != nil {
			fmt.Println("rpc-client need depend balancer")
		}
		// check discover
		if b.discoverBuilder != nil {
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

}
