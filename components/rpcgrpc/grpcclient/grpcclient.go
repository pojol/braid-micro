// 实现文件 基于 grpc 实现的 rpc-client
package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/internal/balancer"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(coordinate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("can't find service node in center")
)

// Client 调用器
type grpcClient struct {
	info meta.ServiceInfo
	parm Parm

	b balancer.IBalancer

	log *blog.Logger

	discoverchan module.IChannel
	linkcache    module.ILinkCache
	ps           module.IPubsub

	connmap sync.Map
}

func BuildWithOption(info meta.ServiceInfo, log *blog.Logger, b balancer.IBalancer, linkcache module.ILinkCache, ps module.IPubsub, opts ...Option) module.IClient {

	p := DefaultClientParm

	for _, opt := range opts {
		opt(&p)
	}

	return &grpcClient{
		info:      info,
		log:       log,
		b:         b,
		linkcache: linkcache,
		ps:        ps,
		parm:      p,
	}
}

func (c *grpcClient) newconn(addr string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var conn *grpc.ClientConn
	var err error

	if len(c.parm.UnaryInterceptors) > 0 {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(c.parm.UnaryInterceptors...)))
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
	c.log.Infof("[braid.client] new connect addr : %v err : %v", addr, err)

	return conn, err
}

func (c *grpcClient) closeconn(conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	doneCh := make(chan error)
	go func() {
		var result error
		if err := conn.Close(); err != nil {
			result = fmt.Errorf("[braid.client] %w %v", err, "failed to close gRPC client")
		}
		doneCh <- result
	}()

	select {
	case <-ctx.Done():
		return errors.New("failed to close gRPC client because of timeout")
	case err := <-doneCh:
		c.log.Infof("[braid.client] close connect addr : %v err : %v", conn.Target(), err)
		return err
	}
}

func (c *grpcClient) Init() error {
	var err error

	c.b.Init() // 初始化自身的负载均衡器
	defer c.b.Run()

	c.discoverchan, err = c.ps.GetTopic(meta.TopicDiscoverServiceUpdate).
		Sub(context.TODO(), meta.ModuleClient+"-"+c.info.ID)
	if err != nil {
		return err
	}

	c.discoverchan.Arrived(func(msg *meta.Message) error {
		dmsg := meta.DecodeUpdateMsg(msg)
		if dmsg.Event == meta.TopicDiscoverServiceNodeAdd {
			_, ok := c.connmap.Load(dmsg.Nod.Address)
			if !ok {
				conn, err := c.newconn(dmsg.Nod.Address)
				if err != nil {
					c.log.Errf("[braid.client] new grpc conn err %s", err.Error())
				} else {
					c.connmap.Store(dmsg.Nod.Address, conn)
				}
			}
		} else if dmsg.Event == meta.TopicDiscoverServiceNodeRmv {
			mc, ok := c.connmap.Load(dmsg.Nod.Address)
			if ok {
				conn := mc.(*grpc.ClientConn)
				err = c.closeconn(conn)
				if err != nil {
					c.log.Errf("[braid.client] close grpc conn err %s", err.Error())
				} else {
					c.connmap.Delete(dmsg.Nod.Address)
				}
			}
		}
		return nil
	})

	return nil
}

func (c *grpcClient) getConn(address string) (*grpc.ClientConn, error) {
	mc, ok := c.connmap.Load(address)
	if !ok {
		return nil, fmt.Errorf("gRPC client Can't find target %s", address)
	}

	conn, ok := mc.(*grpc.ClientConn)
	if !ok {
		return nil, fmt.Errorf("gRPC client failed address : %s", address)
	}

	if conn.GetState() == connectivity.TransientFailure {
		c.log.Warnf("[braid.client] reset connect backoff")
		conn.ResetConnectBackoff()
	}

	return conn, nil
}

func (c *grpcClient) pick(nodName string, token string, link bool) (meta.Node, error) {

	var nod meta.Node
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
	var nod meta.Node

	if (c.linkcache != nil) && token != "" {
		address, _ = c.linkcache.Target(ctx, token, target)
	}

	if address == "" {
		nod, err = c.pick(target, token, c.linkcache != nil)
		if err != nil {
			c.log.Warnf("[braid.client] pick warning %s", err.Error())
			return ""
		}

		address = nod.Address
		if (c.linkcache != nil) && token != "" {
			err = c.linkcache.Link(ctx, token, nod)
			if err != nil {
				c.log.Warnf("[braid.client] link warning %s %s %s", token, target, err.Error())
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
		c.log.Warnf("[braid.client] client get conn warning %s", err.Error())
		return err
	}

	if len(opts) != 0 {
		for _, v := range opts {
			callopt, ok := v.(grpc.CallOption)
			if !ok {
				c.log.Warnf("[braid.client] call option type mismatch")
			}
			grpcopts = append(grpcopts, callopt)
		}
	}

	err = conn.Invoke(ctx, methon, args, reply, grpcopts...)
	if err != nil {
		c.log.Warnf("[braid.client] invoke warning %s, target = %s, methon = %s, addr = %s, token = %s", err.Error(), nodName, methon, address, token)
		if c.linkcache != nil {
			c.linkcache.Unlink(ctx, token)
		}
	}

	return err
}

func (s *grpcClient) Name() string {
	return "GrpcClient"
}

func (c *grpcClient) Run() {

}

func (c *grpcClient) Close() {
	c.discoverchan.Close()
	c.b.Close()
}
