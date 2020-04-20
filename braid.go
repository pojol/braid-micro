package braid

import (
	"context"
	"errors"

	"github.com/pojol/braid/balancer"
	"github.com/pojol/braid/cache/redis"
	"github.com/pojol/braid/caller"
	"github.com/pojol/braid/caller/brpc"
	"github.com/pojol/braid/discover"
	"github.com/pojol/braid/election"
	"github.com/pojol/braid/link"
	"github.com/pojol/braid/log"
	"github.com/pojol/braid/service"
	"github.com/pojol/braid/tracer"
)

const (
	// Balancer 负载均衡
	Balancer = "balancer"

	// Logger 日志
	Logger = "logger"

	// Redis redis
	Redis = "redis"

	// Linker 链接器
	Linker = "linker"

	// Service 服务管理器
	Service = "service"

	// Caller 远程调用RPC
	Caller = "caller"

	// Discover 探索发现节点
	Discover = "discover"

	// Election 选举器
	Election = "election"

	// Tracer 链路追踪
	Tracer = "tracer"
)

type (
	// Box 框架
	Box struct {
		Nodes map[string]interface{}
	}

	// Node box框架中的功能节点
	Node interface {
		Init(interface{}) error
		Run()
		Close()
	}

	// NodeCompose 节点
	NodeCompose struct {
		Ty  string
		Cfg interface{}
	}
)

var (
	b *Box
)

func appendNode(name string, nod interface{}) {

	if _, ok := b.Nodes[name]; !ok {
		b.Nodes[name] = nod
	}

}

// Compose 编排工具
func Compose(nlist []NodeCompose) error {

	// 构造
	b = &Box{
		Nodes: make(map[string]interface{}),
	}

	for _, v := range nlist {

		if v.Ty == Linker {
			l := link.New()
			err := l.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Linker, l)
		} else if v.Ty == Redis {
			r := redis.New()
			err := r.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Redis, r)
		} else if v.Ty == Balancer {
			ba := balancer.New()
			err := ba.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Balancer, ba)
		} else if v.Ty == Caller {
			ca := caller.New()
			err := ca.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Caller, ca)
		} else if v.Ty == Discover {
			di := discover.New()
			err := di.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Discover, di)
		} else if v.Ty == Logger {
			lo := log.New()
			err := lo.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Logger, lo)
		} else if v.Ty == Election {
			el := election.New()
			err := el.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Election, el)
		} else if v.Ty == Tracer {
			tr := tracer.New()
			err := tr.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Tracer, tr)
		} else if v.Ty == Service {
			se := service.New()
			err := se.Init(v.Cfg)
			if err != nil {
				return err
			}
			appendNode(Service, se)
		}

	}

	return nil
}

// Regist 注册服务
func Regist(serviceName string, fc service.ServiceFunc) {
	if _, ok := b.Nodes[Service]; ok {
		s := b.Nodes[Service].(*service.Service)
		s.Regist(serviceName, fc)
	} else {
		panic(errors.New("No subscription service nod"))
	}
}

// Call 远程调用
func Call(parentCtx context.Context, boxName string, serviceName string, token string, body []byte) (res *brpc.RouteRes, err error) {
	if _, ok := b.Nodes[Caller]; ok {
		c := b.Nodes[Caller].(*caller.Caller)
		return c.Call(parentCtx, boxName, serviceName, token, body)
	}

	panic(errors.New("No subscription service nod"))
}

// Run 运行box
func Run() {

	for _, v := range b.Nodes {
		v.(Node).Run()
	}

}

// Close 释放box
func Close() {
	for _, v := range b.Nodes {
		v.(Node).Close()
	}
}
