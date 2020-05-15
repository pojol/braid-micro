package braid

import (
	"context"
	"errors"
	"strconv"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/rpc/dispatcher"
	"github.com/pojol/braid/rpc/register"
	"github.com/pojol/braid/service/balancer"
	"github.com/pojol/braid/service/discover"
	"github.com/pojol/braid/service/election"
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
	// 零时弃用
	// Linker = "linker"

	// Register rpc服务
	Register = "register"

	// Dispatcher 远程调用RPC
	Dispatcher = "dispatcher"

	// Discover 探索发现节点
	Discover = "discover"

	// Election 选举器
	Election = "election"

	// Tracer 链路追踪
	Tracer = "tracer"
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

	Install []string `yaml:"install"`

	Config struct {
		LogPath   string `yaml:"logger_path"`
		LogSuffex string `yaml:"logger_suffex"`

		RegisterListenPort int `yaml:"register_listen_port"`

		ElectionLockTick   int `yaml:"election_lock_tick"`
		ElectionRefushTick int `yaml:"election_refush_tick"`

		RPCDiscoverInterval int `yaml:"dispatcher_discover_interval"`
		RPCPoolInitNum      int `yaml:"dispatcher_pool_init_num"`
		RPCPoolCap          int `yaml:"dispatcher_pool_cap"`
		RPCPoolIdle         int `yaml:"dispatcher_pool_idle"`

		TracerProbabilistic float64 `yaml:"tracer_probabilistic"`
		TracerSlowReq       int     `yaml:"tracer_slow_req"`
		TracerSlowSpan      int     `yaml:"tracer_slow_span"`
	}
}

type (
	// Braid 框架
	Braid struct {
		Nodes map[string]interface{}
	}

	// Node braid框架中的功能节点
	Node interface {
		Init() error
		Run()
		Close()
	}
)

var (
	b *Braid

	// ErrNotDefineName 必须要有一个节点名称
	ErrNotDefineName = errors.New("not define name")
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

	if compose.Name == "" {
		return ErrNotDefineName
	}

	for _, v := range compose.Install {

		if v == Logger {
			lo := log.New(compose.Config.LogPath)
			err := lo.Init()
			if err != nil {
				return err
			}

			nods = append(nods, Logger)
			AppendNode(Logger, lo)
		}

		if v == Register {

			opts := []register.Option{}
			if compose.Tracing {
				opts = append(opts, register.WithTracing())
			}
			if compose.Config.RegisterListenPort != 0 {
				opts = append(opts, register.WithListen(":"+strconv.Itoa(compose.Config.RegisterListenPort)))
			}

			reg := register.New(compose.Name, opts...)
			err := reg.Init()
			if err != nil {
				return err
			}
			nods = append(nods, Register)
			AppendNode(Register, reg)
		}

		if v == Dispatcher {
			dcOpts := []discover.Option{}
			if compose.Config.RPCDiscoverInterval != 0 {
				dcOpts = append(dcOpts, discover.WithInterval(compose.Config.RPCDiscoverInterval))
			}

			dis := discover.New(compose.Name, depend.Consul, dcOpts...)
			err := dis.Init()
			if err != nil {
				return err
			}
			nods = append(nods, Discover)
			AppendNode(Discover, dis)

			ba := balancer.New()
			err = ba.Init()
			if err != nil {
				return err
			}
			nods = append(nods, Balancer)
			AppendNode(Balancer, ba)

			dispOpts := []dispatcher.Option{}
			if compose.Tracing {
				dispOpts = append(dispOpts, dispatcher.WithTracing())
			}
			if compose.Config.RPCPoolCap != 0 {
				dispOpts = append(dispOpts, dispatcher.WithPoolCapacity(compose.Config.RPCPoolCap))
			}
			if compose.Config.RPCPoolInitNum != 0 {
				dispOpts = append(dispOpts, dispatcher.WithPoolInitNum(compose.Config.RPCPoolInitNum))
			}
			if compose.Config.RPCPoolIdle != 0 {
				dispOpts = append(dispOpts, dispatcher.WithPoolIdle(compose.Config.RPCPoolIdle))
			}

			r := dispatcher.New(depend.Consul, dispOpts...)
			err = r.Init()
			if err != nil {
				return err
			}
			nods = append(nods, Dispatcher)
			AppendNode(Dispatcher, r)
		}

		if v == Election {

			opts := []election.Option{}
			if compose.Config.ElectionLockTick != 0 {
				opts = append(opts, election.WithLockTick(compose.Config.ElectionLockTick))
			}
			if compose.Config.ElectionRefushTick != 0 {
				opts = append(opts, election.WithRefushTick(compose.Config.ElectionRefushTick))
			}

			el := election.New(
				compose.Name,
				depend.Consul,
				opts...,
			)

			err := el.Init()
			if err != nil {
				return err
			}
			nods = append(nods, Election)
			AppendNode(Election, el)
		}

		if v == Tracer && compose.Tracing {

			opts := []tracer.Option{}
			if compose.Config.TracerProbabilistic != 0 {
				opts = append(opts, tracer.WithProbabilistic(compose.Config.TracerProbabilistic))
			}
			if compose.Config.TracerSlowReq != 0 {
				opts = append(opts, tracer.WithSlowRequest(compose.Config.TracerSlowReq))
			}
			if compose.Config.TracerSlowSpan != 0 {
				opts = append(opts, tracer.WithSlowSpan(compose.Config.TracerSlowSpan))
			}

			tr := tracer.New(compose.Name, depend.Jaeger, opts...)
			err := tr.Init()
			if err != nil {
				return err
			}
			nods = append(nods, Tracer)
			AppendNode(Tracer, tr)
		}

	}

	log.SysCompose(nods, "braid compose install ")
	return nil
}

// Regist 注册服务
func Regist(serviceName string, fc register.RPCFunc) {
	if _, ok := b.Nodes[Register]; ok {
		s := b.Nodes[Register].(*register.Register)
		s.Regist(serviceName, fc)
	} else {
		panic(errors.New("No subscription service module"))
	}
}

// Call 远程调用
func Call(parentCtx context.Context, nodeName string, serviceName string, token string, body []byte) (out []byte, err error) {
	if _, ok := b.Nodes[Dispatcher]; ok {
		c := b.Nodes[Dispatcher].(dispatcher.IDispatcher)
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
