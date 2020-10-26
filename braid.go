package braid

import (
	"context"
	"fmt"
	"runtime"

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
	log, err := zlb.Build(logger.ERROR)
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

	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			stack := make([]byte, 4<<10) // 4kb
			length := runtime.Stack(stack, true)
			b.logger.Errorf("[PANIC RECOVER] %v %s", err, stack[:length])
		}
	}()

	for _, m := range b.modules {
		m.Run()
	}

}

// Invoke iclient.invoke
func Invoke(ctx context.Context, nodeName, methon, token string, args, reply interface{}) {
	if braidGlobal != nil && braidGlobal.client != nil {
		braidGlobal.client.Invoke(ctx, nodeName, methon, token, args, reply)
	}
}

// Server iserver.server
func Server() interface{} {
	if braidGlobal != nil && braidGlobal.server != nil {
		return braidGlobal.server.Server()
	}

	return nil
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
