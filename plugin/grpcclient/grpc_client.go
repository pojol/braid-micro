package grpcclient

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pojol/braid/internal/pool"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/tracer"
	"google.golang.org/grpc"
)

var (

	// Name client plugin name
	Name = "GRPCClient"

	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(coordinate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("Can't find service node in center")
)

type grpcClientBuilder struct {
	opts []interface{}
}

func newGRPCClient() client.Builder {
	return &grpcClientBuilder{}
}

func (b *grpcClientBuilder) Name() string {
	return Name
}

func (b *grpcClientBuilder) AddOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *grpcClientBuilder) Build(serviceName string, logger logger.ILogger) (client.IClient, error) {

	p := Parm{
		PoolInitNum:  128,
		PoolCapacity: 1024,
		PoolIdle:     time.Second * 120,
	}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	c := &grpcClient{
		serviceName: serviceName,
		parm:        p,
		logger:      logger,
	}

	return c, nil
}

// Client 调用器
type grpcClient struct {
	serviceName string
	parm        Parm
	logger      logger.ILogger

	poolMgr sync.Map
}

func (c *grpcClient) Init() {

}

func (c *grpcClient) Run() {

}

func (c *grpcClient) getConn(address string) (*pool.ClientConn, error) {
	var caConn *pool.ClientConn
	var caPool *pool.GRPCPool

	caPool, err := c.pool(address)
	if err != nil {
		c.logger.Debugf("get rpc pool err %s", err.Error())
		return nil, err
	}

	connCtx, connCancel := context.WithTimeout(context.Background(), time.Second)
	defer connCancel()
	caConn, err = caPool.Get(connCtx)
	if err != nil {
		c.logger.Debugf("get conn by rpc pool err %s", err.Error())
		return nil, err
	}

	return caConn, nil
}

func pick(nodName string, token string, link bool) (discover.Node, error) {

	var nod discover.Node
	var err error

	if token == "" && link {
		nod, err = balancer.Get(nodName).Random()
	} else {
		nod, err = balancer.Get(nodName).Pick()
	}

	if err != nil {
		return nod, err
	}

	if nod.Address == "" {
		return nod, errors.New("address is not available")
	}

	return nod, nil
}

func (c *grpcClient) findTarget(ctx context.Context, token string, target string) string {
	var address string
	var err error
	var nod discover.Node

	if c.parm.byLink && token != "" {
		address, err = c.parm.linker.Target(token, target)
		if err != nil {
			c.logger.Debugf("linker.target warning %s", err.Error())
			return ""
		}
	}

	if address == "" {
		nod, err = pick(target, token, c.parm.byLink)
		if err != nil {
			c.logger.Debugf("pick warning %s", err.Error())
			return ""
		}

		address = nod.Address
		if c.parm.byLink && token != "" {
			llt := tracer.RedisTracer{
				Cmd: "linker.link",
			}
			llt.Begin(ctx)
			err = c.parm.linker.Link(token, nod)
			llt.End()
			if err != nil {
				c.logger.Debugf("link warning %s %s", token, err.Error())
			}
		}
	}

	return address
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

	select {
	case <-ctx.Done():
		return
	default:
	}

	address = c.findTarget(ctx, token, nodName)
	if address == "" {
		c.logger.Debugf("find target warning %s %s", token, nodName)
		return
	}

	conn, err := c.getConn(address)
	if err != nil {
		c.logger.Debugf("client get conn warning %s", err.Error())
		return
	}
	defer conn.Put()

	//opts...
	err = conn.ClientConn.Invoke(ctx, methon, args, reply)
	if err != nil {
		c.logger.Debugf("client invoke warning %s, target = %s, token = %s", err.Error(), nodName, token)
		if c.parm.byLink {
			c.parm.linker.Unlink(token, nodName)
		}

		conn.Unhealthy()
	}

}

// Pool 获取grpc连接池
func (c *grpcClient) pool(address string) (p *pool.GRPCPool, err error) {

	factory := func() (*grpc.ClientConn, error) {
		var conn *grpc.ClientConn
		var err error

		if c.parm.isTracing {
			interceptor := tracer.ClientInterceptor(opentracing.GlobalTracer())
			conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(interceptor))
		} else {
			conn, err = grpc.Dial(address, grpc.WithInsecure())
		}

		if err != nil {
			c.logger.Debugf("rpc pool factory err %s", err.Error())
			return nil, err
		}

		return conn, nil
	}

	pi, ok := c.poolMgr.Load(address)
	if !ok {
		p, err = pool.NewGRPCPool(factory, c.parm.PoolInitNum, c.parm.PoolCapacity, c.parm.PoolIdle)
		if err != nil {
			c.logger.Debugf("new grpc pool err %s", err.Error())
			goto EXT
		}

		c.poolMgr.Store(address, p)
		pi = p
	}

	p = pi.(*pool.GRPCPool)

EXT:
	return p, err
}

func (c *grpcClient) Close() {

}

func init() {
	client.Register(newGRPCClient())
}
