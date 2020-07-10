package braid

import (
	"fmt"

	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linker"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/tracer"
	"github.com/pojol/braid/plugin/balancer"
)

// Braid framework instance
type Braid struct {
	cfg config

	clientBuilder client.Builder
	client        client.IClient

	discoverBuilder discover.Builder
	discover        discover.IDiscover

	linkerBuilder linker.Builder
	linker        linker.ILinker

	electorBuild elector.Builder

	balancerBuilder balancer.Builder

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

	// build
	if b.discoverBuilder != nil {
		b.discover = b.discoverBuilder.Build()
	}

	if b.balancerBuilder != nil {
		balancer.NewGroup(b.balancerBuilder)
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

}

// Client grpc-client
func Client() client.IClient {
	return braidGlobal.client
}

// Close 关闭braid
func (b *Braid) Close() {

	if b.discover != nil {
		b.discover.Close()
	}

}
