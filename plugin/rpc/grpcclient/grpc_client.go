package grpcclient

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/plugin/balancer"

	"github.com/opentracing/opentracing-go"
	"github.com/pojol/braid/internal/pool"
	"github.com/pojol/braid/module/linker"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/tracer"
	"google.golang.org/grpc"
)

type grpcClientBuilder struct {
	cfg Config
}

func newGRPCClient() client.Builder {
	return &grpcClientBuilder{}
}

func (b *grpcClientBuilder) Name() string {
	return ClientName
}

func (b *grpcClientBuilder) SetCfg(cfg interface{}) error {
	cecfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	b.cfg = cecfg
	return nil
}

func (b *grpcClientBuilder) Build(link linker.ILinker, tracing bool) client.IClient {
	c := &grpcClient{
		cfg:       b.cfg,
		linker:    link,
		isTracing: tracing,
	}

	return c
}

// Client 调用器
type grpcClient struct {
	cfg Config

	refushTick *time.Ticker

	linker    linker.ILinker
	isTracing bool

	poolMgr sync.Map
	sync.Mutex
}

var (

	// ClientName client plugin name
	ClientName = "GRPCClient"

	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(coordinate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("Can't find service node in center")
)

func (c *grpcClient) getConn(address string) (*pool.ClientConn, error) {
	var caConn *pool.ClientConn
	var caPool *pool.GRPCPool

	caPool, err := c.pool(address)
	if err != nil {
		return nil, err
	}

	connCtx, connCancel := context.WithTimeout(context.Background(), time.Second)
	defer connCancel()
	caConn, err = caPool.Get(connCtx)
	if err != nil {
		return nil, err
	}

	return caConn, nil
}

func pick(nodName string) (balancer.Node, error) {
	nod, err := balancer.Get(nodName).Pick()
	if err != nil {
		// err log
		return nod, err
	}

	if nod.Address == "" {
		// err log
		return nod, errors.New("address is not available")
	}

	return nod, nil
}

func (c *grpcClient) linked() bool {
	return c.linker != nil
}

func (c *grpcClient) tracing() bool {
	return (c.isTracing == true)
}

// Invoke 执行远程调用
// ctx 链路的上下文，主要用于tracing
// nodName 逻辑节点名称, 用于查找目标节点地址
// methon 方法名，用于定位到具体的rpc 执行函数
// token 用户身份id
// args request
// reply result
func (c *grpcClient) Invoke(ctx context.Context, nodName, methon, token string, args, reply interface{}) {

	var address string
	var err error
	var nod balancer.Node

	if c.linked() {
		address, err = c.linker.Target(token)
		if err != nil {
			// log
			return
		}
	}

	if address == "" {
		nod, err = pick(nodName)
		if err != nil {
			return
		}

		address = nod.Address
		if c.linked() {
			c.linker.Link(token, nod.ID, nod.Address)
		}
	}

	conn, err := c.getConn(address)
	if err != nil {
		//log
		return
	}

	//opts...
	err = conn.ClientConn.Invoke(ctx, methon, args, reply)
	if err != nil {
		if c.linked() {
			c.linker.Unlink(token)
		}

		conn.Unhealthy()
	}

	conn.Put()
}

// Pool 获取grpc连接池
func (c *grpcClient) pool(address string) (p *pool.GRPCPool, err error) {

	factory := func() (*grpc.ClientConn, error) {
		var conn *grpc.ClientConn
		var err error

		if c.tracing() {
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
		log.SysError("rpcClient", "pool", err.Error())
	}

	return p, err
}

func init() {
	client.Register(newGRPCClient())
}