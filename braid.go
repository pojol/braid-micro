package braid

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/service/balancer"
	"github.com/pojol/braid/service/discover"
	"github.com/pojol/braid/service/election"
	"github.com/pojol/braid/service/register"
	"github.com/pojol/braid/service/rpc"
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

	// RPC 远程调用RPC
	RPC = "rpc"

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

	Install []string `yaml:"install"`

	Config struct {
		LogPath   string `yaml:"logger_path"`
		LogSuffex string `yaml:"logger_suffex"`

		RegisterListenPort int `yaml:"register_listen_port"`

		ElectionLockTick   int `yaml:"election_lock_tick"`
		ElectionRefushTick int `yaml:"election_refush_tick"`

		RPCDiscoverInterval int `yaml:"rpc_discover_interval"`
		RPCPoolInitNum      int `yaml:"rpc_pool_init_num"`
		RPCPoolCap          int `yaml:"rpc_pool_cap"`
		RPCPoolIdle         int `yaml:"rpc_pool_idle"`

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

	for _, v := range compose.Install {

		if v == Logger {
			lo := log.New()
			lc := log.Config{
				Mode:   compose.Mode,
				Path:   log.DefaultConfig.Path,
				Suffex: log.DefaultConfig.Suffex,
			}

			if compose.Config.LogPath != "" {
				lc.Path = compose.Config.LogPath
			}
			if compose.Config.LogSuffex != "" {
				lc.Suffex = compose.Config.LogSuffex
			}

			err := lo.Init(lc)
			if err != nil {
				return err
			}

			nods = append(nods, Logger)
			AppendNode(Logger, lo)
		}

		if v == Register {
			reg := register.New()
			regc := register.Config{
				Name:          compose.Name,
				Tracing:       compose.Tracing,
				ListenAddress: register.DefaultConfig.ListenAddress,
			}

			if compose.Config.RegisterListenPort != 0 {
				regc.ListenAddress = ":" + strconv.Itoa(compose.Config.RegisterListenPort)
			}

			err := reg.Init(regc)
			if err != nil {
				return err
			}
			nods = append(nods, Register)
			AppendNode(Register, reg)
		}

		if v == RPC {
			dis := discover.New()
			disc := discover.Config{
				Name:          compose.Name,
				Interval:      discover.DefaultConfg.Interval,
				ConsulAddress: depend.Consul,
			}

			if compose.Config.RPCDiscoverInterval != 0 {
				disc.Interval = compose.Config.RPCDiscoverInterval
			}
			err := dis.Init(disc)
			if err != nil {
				return err
			}
			nods = append(nods, Discover)
			AppendNode(Discover, dis)

			ba := balancer.New()
			err = ba.Init(balancer.Cfg{})
			if err != nil {
				return err
			}
			nods = append(nods, Balancer)
			AppendNode(Balancer, ba)

			r := rpc.New()
			rc := rpc.Config{
				ConsulAddress: depend.Consul,
				Tracing:       compose.Tracing,
				PoolInitNum:   rpc.DefaultConfig.PoolInitNum,
				PoolCapacity:  rpc.DefaultConfig.PoolCapacity,
				PoolIdle:      rpc.DefaultConfig.PoolIdle,
			}

			if compose.Config.RPCPoolCap != 0 {
				rc.PoolCapacity = compose.Config.RPCPoolCap
			}
			if compose.Config.RPCPoolInitNum != 0 {
				rc.PoolInitNum = compose.Config.RPCPoolInitNum
			}
			if compose.Config.RPCPoolIdle != 0 {
				rc.PoolIdle = time.Duration(compose.Config.RPCPoolIdle) * time.Second
			}

			err = r.Init(rc)
			if err != nil {
				return err
			}
			nods = append(nods, RPC)
			AppendNode(RPC, r)
		}

		if v == Election {
			el := election.New()
			elc := election.Config{
				Address:           depend.Consul,
				Name:              compose.Name,
				LockTick:          election.DefaultConfig.LockTick,
				RefushSessionTick: election.DefaultConfig.RefushSessionTick,
			}

			if compose.Config.ElectionLockTick != 0 {
				elc.LockTick = time.Duration(compose.Config.ElectionLockTick) * time.Millisecond
				elc.RefushSessionTick = time.Duration(compose.Config.ElectionRefushTick) * time.Millisecond
			}

			err := el.Init(elc)
			if err != nil {
				return err
			}
			nods = append(nods, Election)
			AppendNode(Election, el)
		}

		if v == Tracer && compose.Tracing {
			tr := tracer.New()
			trc := tracer.Config{
				Endpoint:      depend.Jaeger,
				Name:          compose.Name,
				Probabilistic: tracer.DefaultTracerConfig.Probabilistic,
				SlowRequest:   tracer.DefaultTracerConfig.SlowRequest,
				SlowSpan:      tracer.DefaultTracerConfig.SlowSpan,
			}

			if compose.Config.TracerProbabilistic != 0 {
				trc.Probabilistic = compose.Config.TracerProbabilistic
			}
			if compose.Config.TracerSlowReq != 0 {
				trc.SlowRequest = time.Duration(compose.Config.TracerSlowReq) * time.Millisecond
			}
			if compose.Config.TracerSlowSpan != 0 {
				trc.SlowSpan = time.Duration(compose.Config.TracerSlowSpan) * time.Millisecond
			}

			err := tr.Init(trc)
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
	if _, ok := b.Nodes[RPC]; ok {
		c := b.Nodes[RPC].(rpc.IRPC)
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
