package caller

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pojol/braid/cache/link"

	"github.com/pojol/braid/log"

	"github.com/opentracing/opentracing-go"

	"github.com/pojol/braid/cache/pool"
	"github.com/pojol/braid/service/balancer"
	"github.com/pojol/braid/service/caller/brpc"
	"github.com/pojol/braid/tracer"
	"google.golang.org/grpc"
)

type (

	// Caller 调用器
	Caller struct {
		cfg Config

		refushTick *time.Ticker

		health bool

		poolMgr sync.Map
		sync.Mutex
	}

	// ICaller caller的抽象接口
	ICaller interface {
		Call(context.Context, string, string, string, []byte) ([]byte, error)
	}

	// Config 调用器配置项
	Config struct {
		ConsulAddress string

		PoolInitNum  int
		PoolCapacity int
		PoolIdle     time.Duration

		Tracing bool
	}
)

var (
	defaultConfig = Config{
		ConsulAddress: "http://127.0.0.1:8500",
		PoolInitNum:   8,
		PoolCapacity:  32,
		PoolIdle:      time.Second * 120,
		Tracing:       false,
	}
	c *Caller

	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(cooridnate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("Can't find service node in center")
)

// New 构建指针
func New() *Caller {
	c = &Caller{}
	return c
}

// Init 通过配置构建调用器
func (c *Caller) Init(cfg interface{}) error {
	callerCfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	if callerCfg.PoolInitNum == 0 {
		callerCfg.PoolInitNum = defaultConfig.PoolInitNum
		callerCfg.PoolCapacity = defaultConfig.PoolCapacity
		callerCfg.PoolIdle = defaultConfig.PoolIdle
	}

	c.cfg = callerCfg

	return nil
}

// Get get caller global pointer
func Get() *Caller {
	return c
}

// Run run
func (c *Caller) Run() {

}

// Close 释放调用器
func (c *Caller) Close() {

}

// Call 执行一次rpc调用
func (c *Caller) Call(parentCtx context.Context, nodName string, serviceName string, token string, body []byte) (out []byte, err error) {

	var address string
	var caPool *pool.GRPCPool
	var caConn *pool.ClientConn
	var caCtx context.Context
	var caCancel context.CancelFunc
	var connCtx context.Context
	var connCancel context.CancelFunc
	var method string
	res := new(brpc.RouteRes)

	c.Lock()
	defer c.Unlock()

	if !c.health {
		return out, ErrServiceNotAvailable
	}

	address, err = c.findNode(parentCtx, nodName, serviceName, token)
	if err != nil {
		goto EXT
	}

	caPool, err = c.pool(address)
	if err != nil {
		goto EXT
	}

	connCtx, connCancel = context.WithTimeout(parentCtx, time.Second)
	defer connCancel()
	caConn, err = caPool.Get(connCtx)
	if err != nil {
		goto EXT
	}
	defer caConn.Put()

	caCtx, caCancel = context.WithTimeout(parentCtx, time.Second)
	defer caCancel()

	method = "/brpc.gateway/routing"
	err = caConn.Invoke(caCtx, method, &brpc.RouteReq{
		Nod:     nodName,
		Service: serviceName,
		ReqBody: body,
	}, res)
	if err != nil {
		caConn.Unhealthy()
		goto EXT
	}

EXT:
	if err != nil {
		log.SysError("caller", "do", err.Error())
	}

	return res.ResBody, err
}

// Find 通过查找器获取目标
func (c *Caller) findNode(parentCtx context.Context, nodName string, serviceName string, key string) (string, error) {
	var address string
	var err error

	if key != "" {

		address, err = link.Get().Target(parentCtx, key)
		if err != nil {
			goto EXT
		}

		if address != "" {
			goto EXT
		}

		nod, err := balancer.GetSelector(nodName).Next()
		if err != nil {
			goto EXT
		}

		address = nod.Address
		link.Get().Link(parentCtx, key, address)
	} else {

		nod, err := balancer.GetSelector(nodName).Next()
		if err != nil {
			goto EXT
		}
		address = nod.Address
	}

EXT:
	if err != nil {
		// log
		log.SysError("caller", "findNode", err.Error())
	}

	return address, err
}

// Pool 获取grpc连接池
func (c *Caller) pool(address string) (p *pool.GRPCPool, err error) {

	factory := func() (*grpc.ClientConn, error) {
		var conn *grpc.ClientConn
		var err error

		if c.cfg.Tracing {
			interceptor := tracer.ClientInterceptor(opentracing.GlobalTracer())
			conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(interceptor))
		} else {
			conn, err = grpc.Dial(address, grpc.WithInsecure())
		}

		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	pi, ok := c.poolMgr.Load(address)
	if !ok {
		p, err = pool.NewGRPCPool(factory, c.cfg.PoolInitNum, c.cfg.PoolCapacity, c.cfg.PoolIdle)
		if err != nil {
			goto EXT
		}

		c.poolMgr.Store(address, p)
		pi = p
	}

	p = pi.(*pool.GRPCPool)

EXT:
	if err != nil {
		log.SysError("caller", "pool", err.Error())
	}

	return p, err
}
