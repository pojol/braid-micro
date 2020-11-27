package braid

import (
	"context"
	"math/rand"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
	"github.com/pojol/braid/modules/grpcclient"
	"github.com/pojol/braid/modules/grpcserver"
	"github.com/pojol/braid/modules/mailboxnsq"
	"github.com/pojol/braid/modules/zaplogger"
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
	server        server.IServer

	tracerBuilder tracer.Builder
	tracer        tracer.ITracer

	mailbox mailbox.IMailbox
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

	rand.Seed(time.Now().UnixNano())

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

// RegistModule 注册模块
func (b *Braid) RegistModule(modules ...Module) error {
	//
	for _, plugin := range modules {
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

	if b.tracerBuilder != nil {
		b.tracer, _ = b.tracerBuilder.Build(b.cfg.Name)
		b.modules = append(b.modules, b.tracer)
	}

	if b.clientBuilder != nil {
		if b.tracer != nil {
			opent, ok := b.tracer.GetTracing().(opentracing.Tracer)
			if ok {
				b.clientBuilder.AddOption(grpcclient.AutoOpenTracing(opent))
			}
		}

		if lc, ok := b.moduleMap[module.TyLinkCache]; ok {
			b.clientBuilder.AddOption(grpcclient.AutoLinkCache(lc.(linkcache.ILinkCache)))
		}
		b.client, _ = b.clientBuilder.Build(b.cfg.Name, b.mailbox, b.logger)
		b.modules = append(b.modules, b.client)
	}

	if b.serverBuilder != nil {
		if b.tracer != nil {
			opent, ok := b.tracer.GetTracing().(opentracing.Tracer)
			if ok {
				b.serverBuilder.AddOption(grpcserver.AutoOpenTracing(opent))
			}
		}

		b.server, _ = b.serverBuilder.Build(b.cfg.Name, b.logger)
		b.modules = append(b.modules, b.server)
	}

	return nil
}

// Init braid init
func (b *Braid) Init() {
	for _, m := range b.modules {
		m.Init()
	}
}

// Run 运行braid
func (b *Braid) Run() {
	/*
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
	*/
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

// GetServer iserver.server
func GetServer() interface{} {
	if braidGlobal != nil && braidGlobal.server != nil {
		return braidGlobal.server.Server()
	}

	return nil
}

// Mailbox pub-sub
func Mailbox() mailbox.IMailbox {
	return braidGlobal.mailbox
}

// Tracer tracing
func Tracer() tracer.ITracer {
	return braidGlobal.tracer
}

// Close 关闭braid
func (b *Braid) Close() {

	for _, m := range b.modules {
		m.Close()
	}

}
