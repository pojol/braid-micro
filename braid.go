package braid

import (
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
	"github.com/pojol/braid/plugin/grpcclient"
	"github.com/pojol/braid/plugin/grpcserver"
	"github.com/pojol/braid/plugin/mailboxnsq"
	"github.com/pojol/braid/plugin/zaplogger"
)

// Braid framework instance
type Braid struct {
	cfg config

	logger logger.ILogger

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
func New(name string, mailboxOpts ...interface{}) (*Braid, error) {

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	for _, opt := range mailboxOpts {
		mbb.AddOption(opt)
	}
	mb, err := mbb.Build(name)
	if err != nil {
		return nil, err
	}

	zlb := logger.GetBuilder(zaplogger.Name)
	log, err := zlb.Build(logger.DEBUG)
	if err != nil {
		return nil, err
	}

	braidGlobal = &Braid{
		cfg: config{
			Name: name,
		},
		mailbox:   mb,
		logger:    log,
		moduleMap: make(map[string]module.IModule),
	}

	return braidGlobal, nil
}

// RegistPlugin 注册插件
func (b *Braid) RegistPlugin(plugins ...Plugin) error {

	//
	for _, plugin := range plugins {
		plugin(braidGlobal)
	}

	// build
	for _, builder := range b.builders {

		m, err := builder.Build(b.cfg.Name, b.mailbox, b.logger)
		if err != nil {
			b.logger.Fatalf("build err %s ,%s", err.Error(), builder.Name())
			continue
		}

		b.moduleMap[builder.Type()] = m
		b.modules = append(b.modules, m)
	}

	if b.clientBuilder != nil {
		if b.tracer != nil {
			b.clientBuilder.AddOption(grpcclient.Tracing())
		}

		if lc, ok := b.moduleMap[module.TyLinkCache]; ok {
			b.clientBuilder.AddOption(grpcclient.LinkCache(lc.(linkcache.ILinkCache)))
		}
		b.client, _ = b.clientBuilder.Build(b.cfg.Name, b.logger)
	}

	if b.serverBuilder != nil {
		if b.tracer != nil {
			b.serverBuilder.AddOption(grpcserver.WithTracing())
		}

		b.server, _ = b.serverBuilder.Build(b.cfg.Name, b.logger)
		b.modules = append(b.modules, b.server)
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
