package braid

import (
	"errors"
	"fmt"
	"sync"

	"github.com/pojol/braid-go/depend"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/tracer"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/module/rpc/server"
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

	// modules
	modules *module.BraidModule

	// depend
	depends *depend.BraidDepend

	rw sync.RWMutex
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

func (b *Braid) RegisterDepend(depends ...depend.Depend) error {

	d := &depend.BraidDepend{}

	for _, opt := range depends {
		opt(d)
	}

	if d.Logger == nil { // 日志是必选项
		d.Logger = blog.BuildWithOption()
	}

	b.depends = d

	return nil
}

func (b *Braid) RegisterModule(modules ...module.Module) error {

	p := &module.BraidModule{
		ServiceName: b.name,
		Depends:     b.depends,
	}

	for _, opt := range modules {
		opt(p)
	}

	b.modules = p

	return nil
}

// Init braid init
func (b *Braid) Init() error {
	var err error

	if b.modules.Iclient != nil {
		err = b.modules.Iclient.Init()
		if err != nil {
			b.depends.Logger.Errf("braid init client err %v", err.Error())
			return err
		}
	}

	if b.modules.Iserver != nil {
		err = b.modules.Iserver.Init()
		if err != nil {
			b.depends.Logger.Errf("braid init server err %v", err.Error())
			return err
		}
	}

	if b.modules.Idiscover != nil {
		err = b.modules.Idiscover.Init()
		if err != nil {
			b.depends.Logger.Errf("braid init discover err %v", err.Error())
			return err
		}
	}

	if b.modules.Ielector != nil {
		err = b.modules.Ielector.Init()
		if err != nil {
			b.depends.Logger.Errf("braid init elector err %v", err.Error())
			return err
		}
	}

	if b.modules.Ilinkcache != nil {
		err = b.modules.Ilinkcache.Init()
		if err != nil {
			b.depends.Logger.Errf("braid init linkcache err %v", err.Error())
			return err
		}
	}

	return err
}

// Run 运行braid
func (b *Braid) Run() {
	fmt.Printf(banner, Version)

	if b.modules.Iserver != nil {
		b.modules.Iserver.Run()
	}

	if b.modules.Idiscover != nil {
		b.modules.Idiscover.Run()
	}

	if b.modules.Ielector != nil {
		b.modules.Ielector.Run()
	}

	if b.modules.Ilinkcache != nil {
		b.modules.Ilinkcache.Run()
	}

}

// GetClient get client interface
func Client() client.IClient {
	if braidGlobal != nil && braidGlobal.modules.Iclient != nil {
		return braidGlobal.modules.Iclient
	}
	return nil
}

func Server() server.IServer {
	if braidGlobal != nil && braidGlobal.modules.Iserver != nil {
		return braidGlobal.modules.Iserver
	}
	return nil
}

// Mailbox pub-sub
func Pubsub() pubsub.IPubsub {
	return braidGlobal.modules.Ipubsub
}

// Tracer tracing
func Tracer() tracer.ITracer {
	return braidGlobal.depends.Tracer
}

// Close 关闭braid
func (b *Braid) Close() {

	if b.modules.Iclient != nil {
		b.modules.Iclient.Close()
	}

	if b.modules.Iserver != nil {
		b.modules.Iserver.Close()
	}

	if b.modules.Idiscover != nil {
		b.modules.Idiscover.Close()
	}

	if b.modules.Ielector != nil {
		b.modules.Ielector.Close()
	}

	if b.modules.Ilinkcache != nil {
		b.modules.Ilinkcache.Close()
	}

}
