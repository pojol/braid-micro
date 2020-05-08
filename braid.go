package braid

import (
	"context"
	"errors"
	"time"

	"github.com/pojol/braid/balancer"
	"github.com/pojol/braid/cache/redis"
	"github.com/pojol/braid/caller"
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

	// DefaultListen 默认的侦听端口，和dockerfile中开放的端口保持一致
	DefaultListen = ":1201"
)

// DependConf braid 需要依赖的服务地址配置
type DependConf struct {
	Consul string `yaml:"consul_addr"`
	Redis  string `yaml:"redis_addr"`
	Jaeger string `yaml:"jaeger_addr"`
}

// ComposeConf braid 编排结构
type ComposeConf struct {
	Name    string `yaml:"name"`
	Mode    string `yaml:"mode"`
	Tracing bool   `yaml:"tracing"`

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
			SlowRequest   int64   `yaml:"slow_req"`
			SlowSpan      int64   `yaml:"slow_span"`
			Probabilistic float64 `yaml:"probabilistic"`
		}
		Service struct {
			Open bool `yaml:"open"`
		}
		Caller struct {
			Open        bool `yaml:"open"`
			PoolInitNum int  `yaml:"pool_init_num"`
			PoolCap     int  `yaml:"pool_cap"`
			PoolIdle    int  `yaml:"pool_idle"`
		}
		Linker struct {
			Open bool `yaml:"open"`
		}
		Balancer struct {
			Open bool `yaml:"open"`
		}
		Election struct {
			Open       bool `yaml:"open"`
			LockTick   int  `yaml:"lock_tick"`   // 循环获取锁的间隔 ms (申请成为master的间隔
			RefushTick int  `yaml:"refush_tick"` // 刷新session的间隔 ms (保活的心跳频率
		}
		Discover struct {
			Open bool `yaml:"open"`
			// Interval 发现节点的刷新间隔 ms
			Interval int `yaml:"interval"`
		}
	}
}

type (
	// Braid 框架
	Braid struct {
		Nodes map[string]interface{}
	}

	// Node braid框架中的功能节点
	Node interface {
		Init(interface{}) error
		Run()
		Close()
	}
)

var (
	b *Braid
)

// AppendNode 将功能节点添加到braid中
func AppendNode(name string, nod interface{}) {

	if _, ok := b.Nodes[name]; !ok {
		b.Nodes[name] = nod
	}

}

// Compose 编排工具
func Compose(compose ComposeConf, depend DependConf) error {

	// 构造
	b = &Braid{
		Nodes: make(map[string]interface{}),
	}

	var nods []string

	if compose.Install.Log.Open {
		lo := log.New()
		err := lo.Init(log.Config{
			Mode:   compose.Mode,
			Path:   compose.Install.Log.Path,
			Suffex: compose.Install.Log.Suffex,
		})
		if err != nil {
			return err
		}

		nods = append(nods, Logger)
		AppendNode(Logger, lo)
	}

	if compose.Install.Redis.Open {
		r := redis.New()
		err := r.Init(redis.Config{
			Address:        depend.Redis,
			ReadTimeOut:    time.Millisecond * time.Duration(compose.Install.Redis.ReadTimeout),
			WriteTimeOut:   time.Millisecond * time.Duration(compose.Install.Redis.WriteTimeout),
			ConnectTimeOut: time.Millisecond * time.Duration(compose.Install.Redis.ConnTimeout),
			IdleTimeout:    time.Millisecond * time.Duration(compose.Install.Redis.IdleTimeout),
			MaxIdle:        compose.Install.Redis.MaxIdle,
			MaxActive:      compose.Install.Redis.MaxActive,
		})
		if err != nil {
			return err
		}
		nods = append(nods, Redis)
		AppendNode(Redis, r)
	}

	if compose.Install.Tracer.Open && compose.Tracing {
		tr := tracer.New()
		err := tr.Init(tracer.Config{
			Endpoint:      depend.Jaeger,
			Name:          compose.Name,
			Probabilistic: compose.Install.Tracer.Probabilistic,
			SlowRequest:   compose.Install.Tracer.SlowRequest,
			SlowSpan:      time.Duration(compose.Install.Tracer.SlowSpan) * time.Millisecond,
		})
		if err != nil {
			return err
		}
		nods = append(nods, Tracer)
		AppendNode(Tracer, tr)
	}

	if compose.Install.Linker.Open {
		l := link.New()
		err := l.Init(link.Config{})
		if err != nil {
			return err
		}
		AppendNode(Linker, l)
	}

	if compose.Install.Balancer.Open {
		ba := balancer.New()
		err := ba.Init(balancer.SelectorCfg{})
		if err != nil {
			return err
		}
		nods = append(nods, Balancer)
		AppendNode(Balancer, ba)
	}

	if compose.Install.Discover.Open {
		di := discover.New()
		err := di.Init(discover.Config{
			ConsulAddress: depend.Consul,
			Interval:      compose.Install.Discover.Interval,
		})
		if err != nil {
			return err
		}
		nods = append(nods, Discover)
		AppendNode(Discover, di)
	}

	if compose.Install.Election.Open {
		el := election.New()
		err := el.Init(election.Config{
			Address:           depend.Consul,
			Name:              compose.Name,
			LockTick:          time.Duration(compose.Install.Election.LockTick) * time.Millisecond,
			RefushSessionTick: time.Duration(compose.Install.Election.RefushTick) * time.Millisecond,
		})
		if err != nil {
			return err
		}
		nods = append(nods, Election)
		AppendNode(Election, el)
	}

	if compose.Install.Caller.Open {
		ca := caller.New()
		err := ca.Init(caller.Config{
			ConsulAddress: depend.Consul,
			PoolInitNum:   compose.Install.Caller.PoolInitNum,
			PoolCapacity:  compose.Install.Caller.PoolCap,
			PoolIdle:      time.Duration(compose.Install.Caller.PoolIdle) * time.Second,
			Tracing:       compose.Tracing,
		})
		if err != nil {
			return err
		}
		nods = append(nods, Caller)
		AppendNode(Caller, ca)
	}

	if compose.Install.Service.Open {
		se := service.New()
		err := se.Init(service.Config{
			Tracing:       compose.Tracing,
			Name:          compose.Name,
			ListenAddress: DefaultListen,
		})
		if err != nil {
			return err
		}
		nods = append(nods, Service)
		AppendNode(Service, se)
	}

	log.SysCompose(nods, "braid compose install ")
	return nil
}

// Regist 注册服务
func Regist(serviceName string, fc service.RPCFunc) {
	if _, ok := b.Nodes[Service]; ok {
		s := b.Nodes[Service].(*service.Service)
		s.Regist(serviceName, fc)
	} else {
		panic(errors.New("No subscription service module"))
	}
}

// Call 远程调用
func Call(parentCtx context.Context, nodeName string, serviceName string, token string, body []byte) (out []byte, err error) {
	if _, ok := b.Nodes[Caller]; ok {
		c := b.Nodes[Caller].(caller.ICaller)
		return c.Call(parentCtx, nodeName, serviceName, token, body)
	}

	panic(errors.New("No subscription caller module"))
}

// IsMaster 当前节点是否为主节点
func IsMaster() (bool, error) {
	if _, ok := b.Nodes[Election]; ok {
		e := b.Nodes[Election].(election.IElection)
		return e.IsMaster(), nil
	}

	return false, errors.New("No subscription election module")
}

// Run 运行
func Run() {

	for _, v := range b.Nodes {
		v.(Node).Run()
	}

}

// Close 释放
func Close() {
	for _, v := range b.Nodes {
		v.(Node).Close()
	}
}
