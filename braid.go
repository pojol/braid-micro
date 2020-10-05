package braid

import (
	"fmt"

	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
	"github.com/pojol/braid/plugin/grpcclient"
	"github.com/pojol/braid/plugin/grpcserver"
	"github.com/pojol/braid/plugin/mailboxnsq"
)

// Braid framework instance
type Braid struct {
	cfg config

	builders []module.Builder

	moduleMap map[string]module.IModule
	modules   []module.IModule

	clientBuilder client.Builder
	client        client.IClient

	serverBuilder server.Builder
	server        server.ISserver

	mailbox mailbox.IMailbox

	tracer *tracer.Tracer
}

var (
	braidGlobal *Braid
)

// New 构建braid
func New(name string, mailboxOpts ...interface{}) *Braid {

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	for _, opt := range mailboxOpts {
		mbb.AddOption(opt)
	}
	mb, err := mbb.Build(name)
	if err != nil {
		return nil
	}

	braidGlobal = &Braid{
		cfg: config{
			Name: name,
		},
		mailbox:   mb,
		moduleMap: make(map[string]module.IModule),
	}

	return braidGlobal
}

// RegistPlugin 注册插件
func (b *Braid) RegistPlugin(plugins ...Plugin) error {

	//
	for _, plugin := range plugins {
		plugin(braidGlobal)
	}

	// build
	for _, builder := range b.builders {

		m, err := builder.Build(b.cfg.Name, b.mailbox)
		if err != nil {
			fmt.Println("build err", builder.Name())
			continue
		}

		b.moduleMap[builder.Type()] = m
		b.modules = append(b.modules, m)
	}

	if b.clientBuilder != nil {
		if b.tracer != nil {
			b.clientBuilder.AddOption(grpcclient.WithTracing())
		}

		if lc, ok := b.moduleMap[module.TyLinkCache]; ok {
			b.clientBuilder.AddOption(grpcclient.WithLinkCache(lc.(linkcache.ILinkCache)))
		}
		b.client, _ = b.clientBuilder.Build(b.cfg.Name)
	}

	if b.serverBuilder != nil {
		if b.tracer != nil {
			b.serverBuilder.AddOption(grpcserver.WithTracing())
		}

		b.server, _ = b.serverBuilder.Build(b.cfg.Name)
	}

	return nil
}

// Run 运行braid
func (b *Braid) Run() {

	for _, m := range b.modules {
		m.Run()
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

// Mailbox pub-sub
func Mailbox() mailbox.IMailbox {
	return braidGlobal.mailbox
}

// Close 关闭braid
func (b *Braid) Close() {

	for _, m := range b.modules {
		m.Close()
	}

}
