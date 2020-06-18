package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pojol/braid/3rd/log"

	"github.com/opentracing/opentracing-go"

	"github.com/pojol/braid/internal/pool"
	"github.com/pojol/braid/module/tracer"
	"github.com/pojol/braid/plugin/balancer"
	"github.com/pojol/braid/plugin/discover"
	"google.golang.org/grpc"
)

type (

	// IClient client的抽象接口
	IClient interface {
		Discover()
		Close()
	}

	// Client 调用器
	Client struct {
		cfg config

		discov discover.IDiscover
		bg     *balancer.Group

		refushTick *time.Ticker

		poolMgr sync.Map
		sync.Mutex
	}
)

var (
	c *Client

	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(coordinate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("Can't find service node in center")
)

// New 构建指针
func New(name string, consulAddress string, opts ...Option) IClient {
	const (
		defaultPoolInitNum  = 8
		defaultPoolCapacity = 32
		defaultPoolIdle     = 120
		defaultTracing      = false
	)

	c = &Client{
		cfg: config{
			ConsulAddress: consulAddress,
			PoolInitNum:   defaultPoolInitNum,
			PoolCapacity:  defaultPoolCapacity,
			PoolIdle:      defaultPoolIdle,
			Tracing:       defaultTracing,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	// 这里后面需要做成可选项
	c.bg = balancer.NewGroup()
	c.discov = discover.New(name, consulAddress, c.bg)

	return c
}

// GetConn 获取rpc client连接
func GetConn(target string) (*pool.ClientConn, error) {
	var caConn *pool.ClientConn
	var caPool *pool.GRPCPool

	c.Lock()
	defer c.Unlock()

	address, err := c.bg.Get(target).Pick()
	if err != nil {
		return nil, err
	}

	caPool, err = c.pool(address)
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

// Pool 获取grpc连接池
func (c *Client) pool(address string) (p *pool.GRPCPool, err error) {

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
		log.SysError("rpcClient", "pool", err.Error())
	}

	return p, err
}

// Discover 执行服务发现逻辑
func (c *Client) Discover() {
	c.discov.Run()
}

// Close 关闭服务发现逻辑
func (c *Client) Close() {
	c.discov.Close()
}
