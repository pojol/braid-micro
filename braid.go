package braid

import (
	"errors"
	"fmt"
	"sync"

	"github.com/pojol/braid-go/depend/balancer"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/pubsub"
	"github.com/pojol/braid-go/depend/redis"
	"github.com/pojol/braid-go/depend/tracer"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/rpc/client"
	"github.com/pojol/braid-go/rpc/server"
)

const (
	// Version of braid-go
	Version = "v1.2.26"

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
)

// Braid framework instance
type Braid struct {
	// service name
	name string

	client client.IClient
	server server.IServer

	// modules
	discoverPtr  discover.IDiscover
	linkcachePtr linkcache.ILinkCache
	electorPtr   elector.IElector

	modules []module.IModule

	// depend
	logPtr  *blog.Logger
	pubsub  pubsub.IPubsub
	redis   *redis.Client
	tracer  tracer.ITracer
	balance balancer.IBalancer

	//
	c client.IClient
	s server.IServer

	sync.RWMutex
}

var (
	braidGlobal *Braid
)

func NewService(name string) (*Braid, error) {

	braidGlobal = &Braid{
		name: name,
	}

	return braidGlobal, nil
}

func (b *Braid) RegisterDepend(log *blog.Logger, r *redis.Client, ps pubsub.IPubsub, t tracer.ITracer) error {

	b.logPtr = log
	b.redis = r
	b.pubsub = ps
	b.tracer = t

	return nil
}

func (b *Braid) RegisterClient(opts ...client.Option) {

	if b.pubsub == nil {
		panic(errors.New("Client module need depend Pubsub"))
	}

	balance := balancer.BuildWithOption(b.name, b.pubsub)
	b.balance = balance

	c := client.BuildWithOption(b.name, b.pubsub, balance, b.linkcachePtr, b.tracer, opts...)

	b.c = c
}

func (b *Braid) RegisterServer(opts ...server.Option) {

	s := server.BuildWithOption(b.name, b.tracer, opts...)
	b.s = s

}

func (b *Braid) RegisterModule(mods ...module.IModule) error {

	b.Lock()
	b.modules = append(b.modules, mods...)
	b.Unlock()

	return nil
}

// Init braid init
func (b *Braid) Init() error {
	var err error

	if b.balance != nil {
		b.balance.Init()
	}

	for _, mod := range b.modules {
		err = mod.Init()
		if err != nil {
			blog.Errf("braid init err %v", err.Error())
			break
		}
	}

	return err
}

// Run 运行braid
func (b *Braid) Run() {
	fmt.Printf(banner, Version)

	if b.balance != nil {
		b.balance.Run()
	}

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

	if b.balance != nil {
		b.balance.Close()
	}

	for _, mod := range b.modules {
		mod.Close()
	}

}
