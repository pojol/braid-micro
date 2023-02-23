package braid

import (
	"errors"
	"fmt"

	"github.com/pojol/braid-go/depend"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/module/modules"
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
	mods *modules.BraidModule

	// depend
	depends *depend.BraidDepend
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

func (b *Braid) RegisterModule(mods ...modules.Module) error {

	p := &modules.BraidModule{
		ServiceName: b.name,
		Depends:     b.depends,
	}

	for _, opt := range mods {
		opt(p)
	}

	b.mods = p

	return nil
}

// Init braid init
func (b *Braid) Init() error {
	var err error

	for _, mod := range b.mods.Mods() {
		err = mod.Init()
		if err != nil {
			b.depends.Logger.Errf("braid init %v err %v", mod.Name(), err.Error())
			return err
		}
	}

	return err
}

// Run 运行braid
func (b *Braid) Run() {
	fmt.Printf(banner, Version)

	for _, mod := range b.mods.Mods() {
		mod.Run()
	}

}

// GetClient get client interface
func Client() client.IClient {
	if braidGlobal != nil && braidGlobal.mods.IClient != nil {
		return braidGlobal.mods.IClient
	}
	return nil
}

func Server() server.IServer {
	if braidGlobal != nil && braidGlobal.mods.IServer != nil {
		return braidGlobal.mods.IServer
	}
	return nil
}

// Mailbox pub-sub
func Pubsub() pubsub.IPubsub {
	if braidGlobal != nil {
		return braidGlobal.mods.Ipubsub
	}
	return nil
}

// Close 关闭braid
func (b *Braid) Close() {

	for _, mod := range b.mods.Mods() {
		mod.Close()
	}

}
