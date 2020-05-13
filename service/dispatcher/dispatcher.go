package dispatcher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pojol/braid/log"

	"github.com/opentracing/opentracing-go"

	"github.com/pojol/braid/cache/pool"
	"github.com/pojol/braid/service/balancer"
	"github.com/pojol/braid/service/dispatcher/bproto"
	"github.com/pojol/braid/tracer"
	"google.golang.org/grpc"
)

type (

	// Dispatcher 调用器
	Dispatcher struct {
		cfg Config

		refushTick *time.Ticker

		poolMgr sync.Map
		sync.Mutex
	}

	// IDispatcher caller的抽象接口
	IDispatcher interface {
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
	// DefaultConfig 默认配置
	DefaultConfig = Config{
		ConsulAddress: "http://127.0.0.1:8500",
		PoolInitNum:   8,
		PoolCapacity:  32,
		PoolIdle:      time.Second * 120,
		Tracing:       false,
	}
	r *Dispatcher

	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(cooridnate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("Can't find service node in center")
)

// New 构建指针
func New() *Dispatcher {
	r = &Dispatcher{}
	return r
}

// Init 通过配置构建调用器
func (r *Dispatcher) Init(cfg interface{}) error {
	rCfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	r.cfg = rCfg

	return nil
}

// Run run
func (r *Dispatcher) Run() {

}

// Close 释放调用器
func (r *Dispatcher) Close() {

}

// Call 执行一次rpc调用
func (r *Dispatcher) Call(parentCtx context.Context, nodName string, serviceName string, meta []*bproto.Header, body []byte) (out []byte, err error) {

	var address string
	var caPool *pool.GRPCPool
	var caConn *pool.ClientConn
	var caCtx context.Context
	var caCancel context.CancelFunc
	var connCtx context.Context
	var connCancel context.CancelFunc
	var method string
	res := new(bproto.RouteRes)

	r.Lock()
	defer r.Unlock()

	address, err = r.findNode(parentCtx, nodName, serviceName, "")
	if err != nil {
		goto EXT
	}

	caPool, err = r.pool(address)
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

	method = "/bproto.listen/routing"
	err = caConn.Invoke(caCtx, method, &bproto.RouteReq{
		Nod:     nodName,
		Service: serviceName,
		Meta:    meta,
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
func (r *Dispatcher) findNode(parentCtx context.Context, nodName string, serviceName string, key string) (string, error) {
	var address string
	var err error
	var nod *balancer.Node

	wb, err := balancer.GetGroup(nodName)
	if err != nil {
		goto EXT
	}

	nod, err = wb.Next()
	if err != nil {
		goto EXT
	}

	address = nod.Address

EXT:
	if err != nil {
		// log
		log.SysError("caller", "findNode", err.Error())
	}

	return address, err
}

// Pool 获取grpc连接池
func (r *Dispatcher) pool(address string) (p *pool.GRPCPool, err error) {

	factory := func() (*grpc.ClientConn, error) {
		var conn *grpc.ClientConn
		var err error

		if r.cfg.Tracing {
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

	pi, ok := r.poolMgr.Load(address)
	if !ok {
		p, err = pool.NewGRPCPool(factory, r.cfg.PoolInitNum, r.cfg.PoolCapacity, r.cfg.PoolIdle)
		if err != nil {
			goto EXT
		}

		r.poolMgr.Store(address, p)
		pi = p
	}

	p = pi.(*pool.GRPCPool)

EXT:
	if err != nil {
		log.SysError("caller", "pool", err.Error())
	}

	return p, err
}
