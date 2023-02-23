// 实现文件 基于 grpc 实现的 rpc-client
package clientgrpc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	opentracing "github.com/opentracing/opentracing-go"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/btracer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/internal/balancer"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (

	// Name client plugin name
	Name = "GRPCClient"

	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(coordinate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("can't find service node in center")
)

func BuildWithOption(name string, log *blog.Logger, ps pubsub.IPubsub, lc linkcache.ILinkCache, opts ...client.Option) client.IClient {

	p := client.Parm{
		PoolInitNum:  8,
		PoolCapacity: 64,
		PoolIdle:     time.Second * 100,
	}

	for _, opt := range opts {
		opt(&p)
	}

	c := &grpcClient{
		serviceName: name,
		parm:        p,
		ps:          ps,
		b:           balancer.BuildWithOption(name, ps),
		linkcache:   lc,
		log:         log,
	}

	if p.Tracer != nil {
		c.tracer = p.Tracer.GetTracing().(opentracing.Tracer)
	}

	return c
}

// Client 调用器
type grpcClient struct {
	serviceName string
	parm        client.Parm

	b balancer.IBalancer

	log *blog.Logger

	tracer    opentracing.Tracer
	linkcache linkcache.ILinkCache

	ps pubsub.IPubsub

	connmap sync.Map
}

func (c *grpcClient) newconn(addr string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var conn *grpc.ClientConn
	var err error

	if c.tracer != nil {
		c.parm.Interceptors = append(c.parm.Interceptors, btracer.ClientInterceptor(c.tracer))
	}

	if len(c.parm.Interceptors) > 0 {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(c.parm.Interceptors...)))
		if err != nil {
			goto EXT
		}
	} else {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure())
		if err != nil {
			goto EXT
		}
	}

EXT:
	c.log.Infof("[Client] new connect addr : %v err : %v", addr, err)

	return conn, err
}

func (c *grpcClient) closeconn(conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	doneCh := make(chan error)
	go func() {
		var result error
		if err := conn.Close(); err != nil {
			result = fmt.Errorf("[Client] %w %v", err, "failed to close gRPC client")
		}
		doneCh <- result
	}()

	select {
	case <-ctx.Done():
		return errors.New("failed to close gRPC client because of timeout")
	case err := <-doneCh:
		c.log.Infof("[Client] close connect addr : %v err : %v", conn.Target(), err)
		return err
	}
}

func (c *grpcClient) Name() string {
	return Name
}

func (c *grpcClient) Init() error {
	var err error

	c.b.Init() // 初始化自身的负载均衡器
	defer c.b.Run()

	serviceUpdate := c.ps.LocalTopic(discover.TopicServiceUpdate).Sub(Name)
	serviceUpdate.Arrived(func(msg *pubsub.Message) {
		dmsg := discover.DecodeUpdateMsg(msg)
		if dmsg.Event == discover.EventAddService {
			_, ok := c.connmap.Load(dmsg.Nod.Address)
			if !ok {
				conn, err := c.newconn(dmsg.Nod.Address)
				if err != nil {
					c.log.Errf("[Client] new grpc conn err %s", err.Error())
				} else {
					c.connmap.Store(dmsg.Nod.Address, conn)
				}
			}
		} else if dmsg.Event == discover.EventRemoveService {
			mc, ok := c.connmap.Load(dmsg.Nod.Address)
			if ok {
				conn := mc.(*grpc.ClientConn)
				err = c.closeconn(conn)
				if err != nil {
					c.log.Errf("[Client] close grpc conn err %s", err.Error())
				} else {
					c.connmap.Delete(dmsg.Nod.Address)
				}
			}
		}
	})

	return nil
}

func (c *grpcClient) getConn(address string) (*grpc.ClientConn, error) {
	mc, ok := c.connmap.Load(address)
	if !ok {
		return nil, fmt.Errorf("gRPC client Can't find targe %s", address)
	}

	conn, ok := mc.(*grpc.ClientConn)
	if !ok {
		return nil, fmt.Errorf("gRPC client failed address : %s", address)
	}

	if conn.GetState() == connectivity.TransientFailure {
		c.log.Warnf("[Client] reset connect backoff")
		conn.ResetConnectBackoff()
	}

	return conn, nil
}

func (c *grpcClient) pick(nodName string, token string, link bool) (service.Node, error) {

	var nod service.Node
	var err error

	if token == "" && link {
		nod, err = c.b.Pick(balancer.StrategyRandom, nodName)
	} else {
		nod, err = c.b.Pick(balancer.StrategySwrr, nodName)
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
	var nod service.Node

	if (c.linkcache != nil) && token != "" {
		address, _ = c.linkcache.Target(token, target)
	}

	if address == "" {
		nod, err = c.pick(target, token, c.linkcache != nil)
		if err != nil {
			c.log.Warnf("[Client] pick warning %s", err.Error())
			return ""
		}

		address = nod.Address
		if (c.linkcache != nil) && token != "" {
			err = c.linkcache.Link(token, nod)
			if err != nil {
				c.log.Warnf("[Client] link warning %s %s %s", token, target, err.Error())
			}
		}
	}

	return address
}

// Invoke grpc call
func (c *grpcClient) Invoke(ctx context.Context, nodName, methon, token string, args, reply interface{}, opts ...interface{}) error {

	var address string
	var grpcopts []grpc.CallOption

	address = c.findTarget(ctx, token, nodName)
	if address == "" {
		return fmt.Errorf("find target warning token : %s node : %s", token, nodName)
	}

	conn, err := c.getConn(address)
	if err != nil {
		c.log.Warnf("[Client] client get conn warning %s", err.Error())
		return err
	}

	if len(opts) != 0 {
		for _, v := range opts {
			callopt, ok := v.(grpc.CallOption)
			if !ok {
				c.log.Warnf("[Client] call option type mismatch")
			}
			grpcopts = append(grpcopts, callopt)
		}
	}

	err = conn.Invoke(ctx, methon, args, reply, grpcopts...)
	if err != nil {
		c.log.Warnf("[Client] invoke warning %s, target = %s, methon = %s, addr = %s, token = %s", err.Error(), nodName, methon, address, token)
		if c.linkcache != nil {
			c.linkcache.Unlink(token)
		}
	}

	return err
}

func (c *grpcClient) Run() {

}

func (c *grpcClient) Close() {
	c.b.Close()
}
