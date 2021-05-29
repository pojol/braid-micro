package braid

import (
	"errors"
	"fmt"
	"sync"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/module/rpc/server"
	"github.com/pojol/braid-go/module/tracer"
	"github.com/pojol/braid-go/modules/balancernormal"
	"github.com/pojol/braid-go/modules/discoverconsul"
	"github.com/pojol/braid-go/modules/electorconsul"
	"github.com/pojol/braid-go/modules/electork8s"
	"github.com/pojol/braid-go/modules/grpcclient"
	"github.com/pojol/braid-go/modules/grpcserver"
	"github.com/pojol/braid-go/modules/jaegertracing"
	"github.com/pojol/braid-go/modules/linkerredis"
	"github.com/pojol/braid-go/modules/moduleparm"
	"github.com/pojol/braid-go/modules/pubsubnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
)

const (
	// Version of braid-go
	Version = "v1.2.25"

	banner = `
 _               _     _ 
| |             (_)   | |
| |__  _ __ __ _ _  __| |
| '_ \| '__/ _' | |/ _' |
| |_) | | | (_| | | (_| |
|_.__/|_|  \__,_|_|\__,_| %s

`
)

var (
	ErrTypeConvFailed = errors.New("type conversion failed")

	// 默认提供的模块
	LoggerZap      = zaplogger.Name
	PubsubNsq      = pubsubnsq.Name
	DiscoverConsul = discoverconsul.Name
	ElectorConsul  = electorconsul.Name
	ElectorK8s     = electork8s.Name
	ClientGRPC     = grpcclient.Name
	ServerGRPC     = grpcserver.Name
	TracerJaeger   = jaegertracing.Name
	BalancerSWRR   = balancernormal.Name
	LinkcacheRedis = linkerredis.Name
)

// Braid framework instance
type Braid struct {
	name string // service name

	logger logger.ILogger

	builderMap map[module.ModuleType]module.IBuilder

	modules []module.IModule

	client    client.IClient
	server    server.IServer
	tracer    tracer.ITracer
	linkcache linkcache.ILinkCache
	balancer  balancer.IBalancer
	pubsub    pubsub.IPubsub

	sync.RWMutex
}

var (
	braidGlobal *Braid
)

func NewService(name string) (*Braid, error) {

	braidGlobal = &Braid{
		name:       name,
		builderMap: make(map[module.ModuleType]module.IBuilder),
	}

	return braidGlobal, nil
}

func Module(name string, opts ...interface{}) module.IBuilder {
	builder := module.GetBuilder(name)
	if builder != nil {
		for _, opt := range opts {
			builder.AddModuleOption(opt)
		}

		return builder
	}

	panic(fmt.Errorf("unknow module %v", name))
}

func (b *Braid) Register(builders ...module.IBuilder) error {
	for _, build := range builders {
		b.builderMap[build.Type()] = build
	}

	// init base module
	if loggerBuilder, ok := b.builderMap[module.Logger]; ok {
		li := loggerBuilder.Build(b.name)
		ilog, t := li.(logger.ILogger)
		if !t {
			panic(ErrTypeConvFailed)
		}
		b.logger = ilog
	} else {
		panic(fmt.Errorf("missing required dependencies => %v", "logger"))
	}

	if pubsubBuilder, ok := b.builderMap[module.Pubsub]; ok {
		pbi := pubsubBuilder.Build(b.name, moduleparm.WithLogger(b.logger))
		ipb, t := pbi.(pubsub.IPubsub)
		if !t {
			panic(ErrTypeConvFailed)
		}
		b.pubsub = ipb
	} else {
		panic(fmt.Errorf("missing required dependencies => %v", "pub-sub"))
	}

	if balancerBuilder, ok := b.builderMap[module.Balancer]; ok {
		bi := balancerBuilder.Build(b.name,
			moduleparm.WithLogger(b.logger),
			moduleparm.WithPubsub(b.pubsub),
		)
		ib, t := bi.(balancer.IBalancer)
		if !t {
			panic(ErrTypeConvFailed)
		}
		b.balancer = ib
		b.modules = append(b.modules, ib)
	}

	if tracerBuilder, ok := b.builderMap[module.Tracer]; ok {
		ti := tracerBuilder.Build(b.name, moduleparm.WithLogger(b.logger))
		it, t := ti.(tracer.ITracer)
		if !t {
			panic(ErrTypeConvFailed)
		}
		b.tracer = it
		b.modules = append(b.modules, it)
	}

	if linkBuilder, ok := b.builderMap[module.Linkcache]; ok {
		li := linkBuilder.Build(b.name,
			moduleparm.WithLogger(b.logger),
			moduleparm.WithPubsub(b.pubsub),
		)
		il, t := li.(linkcache.ILinkCache)
		if !t {
			panic(ErrTypeConvFailed)
		}
		b.linkcache = il
		b.modules = append(b.modules, il)
	}

	if discoverBuilder, ok := b.builderMap[module.Discover]; ok {
		di := discoverBuilder.Build(b.name,
			moduleparm.WithLogger(b.logger),
			moduleparm.WithPubsub(b.pubsub),
		)
		id, t := di.(discover.IDiscover)
		if !t {
			panic(ErrTypeConvFailed)
		}

		b.modules = append(b.modules, id)
	}

	if electorBuilder, ok := b.builderMap[module.Elector]; ok {
		ei := electorBuilder.Build(b.name,
			moduleparm.WithLogger(b.logger),
			moduleparm.WithPubsub(b.pubsub),
		)
		ie, t := ei.(elector.IElector)
		if !t {
			panic(ErrTypeConvFailed)
		}
		b.modules = append(b.modules, ie)
	}

	// init function module
	if clientBuilder, ok := b.builderMap[module.Client]; ok {

		baseOpts := []moduleparm.Option{}
		baseOpts = append(baseOpts, moduleparm.WithLogger(b.logger))
		baseOpts = append(baseOpts, moduleparm.WithPubsub(b.pubsub))
		if b.tracer != nil {
			baseOpts = append(baseOpts, moduleparm.WithTracer(b.tracer))
		}
		if b.linkcache != nil {
			baseOpts = append(baseOpts, moduleparm.WithLinkcache(b.linkcache))
		}
		if b.balancer != nil {
			baseOpts = append(baseOpts, moduleparm.WithBalancer(b.balancer))
		} else {
			panic(fmt.Errorf("missing required dependencies => %v", "balancer"))
		}

		islice := make([]interface{}, len(baseOpts))
		for k, v := range baseOpts {
			islice[k] = v
		}
		ci := clientBuilder.Build(b.name, islice...)
		ic, t := ci.(client.IClient)
		if !t {
			panic(ErrTypeConvFailed)
		}
		b.client = ic
		b.modules = append(b.modules, ic)
	}

	if serverBuilder, ok := b.builderMap[module.Server]; ok {

		baseOpts := []moduleparm.Option{}
		baseOpts = append(baseOpts, moduleparm.WithLogger(b.logger))
		if b.tracer != nil {
			baseOpts = append(baseOpts, moduleparm.WithTracer(b.tracer))
		}

		islice := make([]interface{}, len(baseOpts))
		for k, v := range baseOpts {
			islice[k] = v
		}
		si := serverBuilder.Build(b.name, islice...)
		is, t := si.(server.IServer)
		if !t {
			panic(ErrTypeConvFailed)
		}
		b.server = is
		b.modules = append(b.modules, is)
	}

	return nil
}

// Init braid init
func (b *Braid) Init() {
	var err error

	for _, mod := range b.modules {
		err = mod.Init()
		if err != nil {
			b.logger.Errorf("braid init err %v", err.Error())
			break
		}
	}

}

// Run 运行braid
func (b *Braid) Run() {
	fmt.Printf(banner, Version)

	for _, mod := range b.modules {
		mod.Run()
	}
}

// GetClient get client interface
func Client() client.IClient {
	if braidGlobal != nil && braidGlobal.client != nil {
		return braidGlobal.client
	}
	return nil
}

func Server() server.IServer {
	return braidGlobal.server
}

// Mailbox pub-sub
func Pubsub() pubsub.IPubsub {
	return braidGlobal.pubsub
}

// Tracer tracing
func Tracer() tracer.ITracer {
	return braidGlobal.tracer
}

// Close 关闭braid
func (b *Braid) Close() {

	for _, mod := range b.modules {
		mod.Close()
	}

}
