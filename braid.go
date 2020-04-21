package braid

import (
	"context"
	"errors"
	"time"

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

// ComposeConf braid 编排结构
type ComposeConf struct {
	Name    string `yaml:"name"`
	Mode    string `yaml:"mode"`
	Tracing bool   `yaml:"tracing"`

	Depend struct {
		Consul string `yaml:"consul_addr"`
		Redis  string `yaml:"redis_addr"`
		Jaeger string `yaml:"jaeger_addr"`
	}

	Install struct {
		Log struct {
			Open   bool   `yaml:"open"`
			Path   string `yaml:"path"`
			Suffex string `yaml:"suffex"`
		}
		Redis struct {
			Open         bool `yaml:"open"`
			ReadTimeout  int  `yaml:"read_timeout"`
			WriteTimeout int  `yaml:"write_timeout"`
			ConnTimeout  int  `yaml:"conn_timeout"`
			IdleTimeout  int  `yaml:"idle_timeout"`
			MaxIdle      int  `yaml:"max_idle"`
			MaxActive    int  `yaml:"max_active"`
		}
		Tracer struct {
			Open          bool    `yaml:"open"`
			Probabilistic float64 `yaml:"probabilistic"`
		}
		Service struct {
			Open bool `yaml:"open"`
		}
		Caller struct {
			Open bool `yaml:"open"`
		}
		Linker struct {
			Open bool `yaml:"open"`
		}
		Balancer struct {
			Open bool `yaml:"open"`
		}
		Election struct {
			Open bool `yaml:"open"`
		}
		Discover struct {
			Open     bool `yaml:"open"`
			Interval int  `yaml:"interval"`
		}
	}
}

type (
	// Braid 框架
	Braid struct {
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
	b *Braid
)

func appendNode(name string, nod interface{}) {

	if _, ok := b.Nodes[name]; !ok {
		b.Nodes[name] = nod
	}

}

// Compose 编排工具
func Compose(conf ComposeConf) error {

	// 构造
	b = &Braid{
		Nodes: make(map[string]interface{}),
	}

	if conf.Install.Log.Open {
		lo := log.New()
		err := lo.Init(log.Config{
			Mode:   conf.Mode,
			Path:   conf.Install.Log.Path,
			Suffex: conf.Install.Log.Suffex,
		})
		if err != nil {
			return err
		}
		appendNode(Logger, lo)
	}

	if conf.Install.Redis.Open {
		r := redis.New()
		err := r.Init(redis.Config{
			Address:        conf.Depend.Redis,
			ReadTimeOut:    time.Millisecond * time.Duration(conf.Install.Redis.ReadTimeout),
			WriteTimeOut:   time.Millisecond * time.Duration(conf.Install.Redis.WriteTimeout),
			ConnectTimeOut: time.Millisecond * time.Duration(conf.Install.Redis.ConnTimeout),
			IdleTimeout:    time.Millisecond * time.Duration(conf.Install.Redis.IdleTimeout),
			MaxIdle:        conf.Install.Redis.MaxIdle,
			MaxActive:      conf.Install.Redis.MaxActive,
		})
		if err != nil {
			return err
		}
		appendNode(Redis, r)
	}

	if conf.Install.Tracer.Open {
		tr := tracer.New()
		err := tr.Init(tracer.Config{
			Endpoint:      conf.Depend.Jaeger,
			Name:          conf.Name,
			Probabilistic: conf.Install.Tracer.Probabilistic,
		})
		if err != nil {
			return err
		}
		appendNode(Tracer, tr)
	}

	if conf.Install.Linker.Open {
		l := link.New()
		err := l.Init(link.Config{})
		if err != nil {
			return err
		}
		appendNode(Linker, l)
	}

	if conf.Install.Balancer.Open {
		ba := balancer.New()
		err := ba.Init(balancer.SelectorCfg{})
		if err != nil {
			return err
		}
		appendNode(Balancer, ba)
	}

	if conf.Install.Discover.Open {
		di := discover.New()
		err := di.Init(discover.Config{
			ConsulAddress: conf.Depend.Consul,
			Interval:      conf.Install.Discover.Interval,
		})
		if err != nil {
			return err
		}
		appendNode(Discover, di)
	}

	if conf.Install.Election.Open {
		el := election.New()
		err := el.Init(election.Config{
			Address: conf.Depend.Consul,
			Name:    conf.Name,
		})
		if err != nil {
			return err
		}
		appendNode(Election, el)
	}

	if conf.Install.Caller.Open {
		ca := caller.New()
		err := ca.Init(caller.Config{
			ConsulAddress: conf.Depend.Consul,
			PoolInitNum:   8,
			PoolCapacity:  32,
			PoolIdle:      time.Second * 120,
			Tracing:       conf.Tracing,
		})
		if err != nil {
			return err
		}
		appendNode(Caller, ca)
	}

	if conf.Install.Service.Open {
		se := service.New()
		err := se.Init(service.Config{
			Tracing: conf.Tracing,
			Name:    conf.Name,
		})
		if err != nil {
			return err
		}
		appendNode(Service, se)
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
