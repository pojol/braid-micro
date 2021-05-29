// 实现文件 基于 grpc 实现的 rpc-client
package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/balancernormal"
	"github.com/pojol/braid-go/modules/jaegertracing"
	"github.com/pojol/braid-go/modules/moduleparm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
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

func newGRPCClient() module.IBuilder {
	return &grpcClientBuilder{}
}

func (b *grpcClientBuilder) Name() string {
	return Name
}

func (b *grpcClientBuilder) Type() module.ModuleType {
	return module.Client
}

func (b *grpcClientBuilder) AddModuleOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *grpcClientBuilder) Build(name string, buildOpts ...interface{}) interface{} {

	bp := moduleparm.BuildParm{}
	for _, opt := range buildOpts {
		opt.(moduleparm.Option)(&bp)
	}

	p := Parm{
		PoolInitNum:  8,
		PoolCapacity: 64,
		PoolIdle:     time.Second * 100,
	}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	if bp.Balancer == nil {
		panic("")
	}

	c := &grpcClient{
		serviceName: name,
		parm:        p,
		logger:      bp.Logger,
		ps:          bp.PS,
		b:           bp.Balancer,
		linkcache:   bp.Linkcache,
	}

	if bp.Tracer != nil {
		c.tracer = bp.Tracer.GetTracing().(opentracing.Tracer)
	}

	return c
}

// Client 调用器
type grpcClient struct {
	serviceName string
	parm        Parm

	b      balancer.IBalancer
	logger logger.ILogger

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
		interceptor := jaegertracing.ClientInterceptor(c.tracer)
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(interceptor))
		if err != nil {
			goto EXT
		}
		c.logger.Debugf("new connect %v tracer open", addr)
	} else {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure())
		if err != nil {
			goto EXT
		}
		c.logger.Debugf("new connect %v tracer close", addr)
	}
	if err != nil {
		err = fmt.Errorf("%w %v", err, "failed to dial to gRPC server")
		goto EXT
	}

EXT:
	return conn, err
}

func (c *grpcClient) closeconn(conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	doneCh := make(chan error)
	go func() {
		var result error
		if err := conn.Close(); err != nil {
			result = fmt.Errorf("%w %v", err, "failed to close gRPC client")
		}
		doneCh <- result
	}()

	select {
	case <-ctx.Done():
		return errors.New("failed to close gRPC client because of timeout")
	case err := <-doneCh:
		return err
	}
}

func (c *grpcClient) Init() error {
	var err error

	serviceUpdate := c.ps.GetTopic(discover.ServiceUpdate).Sub(Name)
	serviceUpdate.Arrived(func(msg *pubsub.Message) {
		dmsg := discover.DecodeUpdateMsg(msg)
		if dmsg.Event == discover.EventAddService {
			_, ok := c.connmap.Load(dmsg.Nod.Address)
			if !ok {
				conn, err := c.newconn(dmsg.Nod.Address)
				if err != nil {
					c.logger.Errorf("new grpc conn err %s", err.Error())
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
					c.logger.Errorf("close grpc conn err %s", err.Error())
				} else {
					c.connmap.Delete(dmsg.Nod.Address)
				}
			}
		}
	})

	return nil
}

func (c *grpcClient) Run() {

}

func (c *grpcClient) getConn(address string) (*grpc.ClientConn, error) {
	mc, ok := c.connmap.Load(address)
	if !ok {
		return nil, errors.New("gRPC client Can't find target")
	}

	conn, ok := mc.(*grpc.ClientConn)
	if !ok {
		return nil, fmt.Errorf("gRPC client failed address : %s", address)
	}

	if conn.GetState() == connectivity.TransientFailure {
		c.logger.Debugf("reset connect backoff")
		conn.ResetConnectBackoff()
	}

	return conn, nil
}

func (c *grpcClient) pick(nodName string, token string, link bool) (discover.Node, error) {

	var nod discover.Node
	var err error

	if token == "" && link {
		nod, err = c.b.Pick(balancernormal.StrategyRandom, nodName)
	} else {
		nod, err = c.b.Pick(balancernormal.StrategySwrr, nodName)
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

	if (c.linkcache != nil) && token != "" {
		address, _ = c.linkcache.Target(token, target)
	}

	if address == "" {
		nod, err = c.pick(target, token, c.linkcache != nil)
		if err != nil {
			c.logger.Debugf("pick warning %s", err.Error())
			return ""
		}

		address = nod.Address
		if (c.linkcache != nil) && token != "" {
			err = c.linkcache.Link(token, nod)
			if err != nil {
				c.logger.Debugf("link warning %s %s %s", token, target, err.Error())
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
		return fmt.Errorf("find target warning %s %s", token, nodName)
	}

	conn, err := c.getConn(address)
	if err != nil {
		c.logger.Debugf("client get conn warning %s", err.Error())
		return err
	}

	if len(opts) != 0 {
		for _, v := range opts {
			callopt, ok := v.(grpc.CallOption)
			if !ok {

			}
			grpcopts = append(grpcopts, callopt)
		}
	}

	err = conn.Invoke(ctx, methon, args, reply, grpcopts...)
	if err != nil {
		c.logger.Warnf("client invoke warning %s, target = %s, methon = %s, addr = %s, token = %s", err.Error(), nodName, methon, address, token)
		if c.linkcache != nil {
			c.linkcache.Unlink(token)
		}
	}

	return err
}

func (c *grpcClient) Close() {
}

func init() {
	module.Register(newGRPCClient())
}
