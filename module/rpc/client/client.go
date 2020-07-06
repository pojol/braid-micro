package client

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/plugin/balancer"
	_ "github.com/pojol/braid/plugin/balancer/swrrbalancer"
	"github.com/pojol/braid/plugin/discover"
	"github.com/pojol/braid/plugin/linker"
	"github.com/pojol/braid/plugin/linker/redislinker"

	"github.com/opentracing/opentracing-go"
	"github.com/pojol/braid/internal/pool"
	"github.com/pojol/braid/module/tracer"
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

		discov        discover.IDiscover
		discovBuilder discover.Builder

		bg     *balancer.Group
		linker linker.ILinker

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

// New 构建 grpc client
// name 节点名
// discoverOpt 发现选项
// opts 可选选项
func New(name string, discoverOpt Option, opts ...Option) IClient {
	const (
		defaultPoolInitNum  = 8
		defaultPoolCapacity = 32
		defaultPoolIdle     = 120
		defaultTracing      = false
	)

	c = &Client{
		cfg: config{
			Name:         name,
			PoolInitNum:  defaultPoolInitNum,
			PoolCapacity: defaultPoolCapacity,
			PoolIdle:     defaultPoolIdle,
			Tracing:      defaultTracing,
		},
	}

	discoverOpt(c)

	for _, opt := range opts {
		opt(c)
	}

	if c.cfg.Link {
		c.linker = linker.GetBuilder(redislinker.LinkerName).Build(nil)
	}

	// 这里后面需要做成可选项
	c.bg = balancer.NewGroup()
	c.discov = c.discovBuilder.Build(c.bg, c.linker)

	return c
}

func getConn(address string) (*pool.ClientConn, error) {
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
	nod, err := c.bg.Get(nodName).Pick()
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

// Invoke 执行远程调用
// ctx 链路的上下文，主要用于tracing
// nodName 逻辑节点名称, 用于查找目标节点地址
// methon 方法名，用于定位到具体的rpc 执行函数
// token 用户身份id
// args request
// reply result
func Invoke(ctx context.Context, nodName, methon string, token string, args, reply interface{}, opts ...grpc.CallOption) {

	var address string
	var err error
	var nod balancer.Node

	if c.cfg.Link && token != "" {
		address, err = c.linker.Target(token)
		if err != nil {
			// err log
			return
		}

		if address == "" {
			nod, err = pick(nodName)
			if err != nil {
				return
			}

			err = c.linker.Link(token, nod.ID, nod.Address)
			if err != nil {
				// err log
				return
			}
			address = nod.Address
		}
	} else {
		nod, err = pick(nodName)
		if err != nil {
			return
		}

		address = nod.Address
	}

	conn, err := getConn(address)
	if err != nil {
		//log
		return
	}

	err = conn.ClientConn.Invoke(ctx, methon, args, reply, opts...)
	if err != nil {
		if c.cfg.Link && token != "" {
			c.linker.Unlink(token)
		}

		conn.Unhealthy()
	}

	conn.Put()
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
	c.discov.Discover()
}

// Close 关闭服务发现逻辑
func (c *Client) Close() {
	c.discov.Close()
}
