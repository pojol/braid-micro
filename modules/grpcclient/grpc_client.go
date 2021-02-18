package grpcclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/modules/balancergroupbase"
	"github.com/pojol/braid-go/modules/balancerrandom"
	"github.com/pojol/braid-go/modules/balancerswrr"
	"github.com/pojol/braid-go/modules/jaegertracing"
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

func newGRPCClient() client.Builder {
	return &grpcClientBuilder{}
}

func (b *grpcClientBuilder) Name() string {
	return Name
}

func (b *grpcClientBuilder) AddOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *grpcClientBuilder) Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (client.IClient, error) {

	p := Parm{
		PoolInitNum:      8,
		PoolCapacity:     64,
		PoolIdle:         time.Second * 100,
		balancerStrategy: []string{balancerrandom.Name, balancerswrr.Name},
		balancerGroup:    balancergroupbase.Name,
	}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	c := &grpcClient{
		serviceName: serviceName,
		parm:        p,
		logger:      logger,
		mb:          mb,
	}

	return c, nil
}

// Client 调用器
type grpcClient struct {
	serviceName string
	parm        Parm
	bg          balancer.IBalancerGroup
	logger      logger.ILogger

	mb                mailbox.IMailbox
	addTargetConsumer mailbox.IConsumer
	rmvTargetConsumer mailbox.IConsumer

	connmap sync.Map
}

func (c *grpcClient) newconn(addr string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var conn *grpc.ClientConn
	var err error

	if c.parm.tracer != nil {
		interceptor := jaegertracing.ClientInterceptor(c.parm.tracer)
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(interceptor))
	} else {
		conn, err = grpc.DialContext(ctx, addr, grpc.WithInsecure())
	}
	if err != nil {
		return nil, fmt.Errorf("%w %v", err, "failed to dial to gRPC server")
	}

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
	bgb := module.GetBuilder(c.parm.balancerGroup)
	bgb.AddOption(balancergroupbase.WithStrategy(c.parm.balancerStrategy))
	bg, err := bgb.Build(c.serviceName, c.mb, c.logger)
	if err != nil {
		return fmt.Errorf("Dependency check error %v [%v]", "balancer", err.Error())
	}

	c.bg = bg.(balancer.IBalancerGroup)
	err = c.bg.Init()
	if err != nil {
		return fmt.Errorf("Dependency check error %v [%v]", "balancer", err.Error())
	}

	c.addTargetConsumer, err = c.mb.Sub(mailbox.Proc, discover.AddService).Shared()
	if err != nil {
		return fmt.Errorf("Dependency check error %v [%v]", "mailbox", discover.AddService)
	}

	c.addTargetConsumer.OnArrived(func(msg mailbox.Message) error {
		nod := discover.Node{}
		json.Unmarshal(msg.Body, &nod)

		_, ok := c.connmap.Load(nod.Address)
		if !ok {
			conn, err := c.newconn(nod.Address)
			if err != nil {
				c.logger.Errorf("new grpc conn err %s", err.Error())
			} else {
				c.connmap.Store(nod.Address, conn)
			}
		}

		return nil
	})

	c.rmvTargetConsumer, err = c.mb.Sub(mailbox.Proc, discover.RmvService).Shared()
	if err != nil {
		return fmt.Errorf("Dependency check error %v [%v]", "mailbox", discover.RmvService)
	}
	c.rmvTargetConsumer.OnArrived(func(msg mailbox.Message) error {

		nod := discover.Node{}
		json.Unmarshal(msg.Body, &nod)

		mc, ok := c.connmap.Load(nod.Address)
		if ok {
			conn := mc.(*grpc.ClientConn)
			err = c.closeconn(conn)
			if err != nil {
				c.logger.Errorf("close grpc conn err %s", err.Error())
			} else {
				c.connmap.Delete(nod.Address)
			}
		}

		return nil
	})

	return nil
}

func (c *grpcClient) Run() {
	c.bg.Run()
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
		conn.ResetConnectBackoff()
	}

	return conn, nil
}

func (c *grpcClient) pick(nodName string, token string, link bool) (discover.Node, error) {

	var nod discover.Node
	var err error

	if token == "" && link {
		nod, err = c.bg.Pick(balancerrandom.Name, nodName)
	} else {
		nod, err = c.bg.Pick(balancerswrr.Name, nodName)
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
		nod, err = c.pick(target, token, c.parm.byLink)
		if err != nil {
			c.logger.Debugf("pick warning %s", err.Error())
			return ""
		}

		address = nod.Address
		if c.parm.byLink && token != "" {
			err = c.parm.linker.Link(token, nod)
			if err != nil {
				c.logger.Debugf("link warning %s %s", token, err.Error())
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
		c.logger.Debugf("client invoke warning %s, target = %s, token = %s", err.Error(), nodName, token)
		if c.parm.byLink {
			c.parm.linker.Unlink(token, nodName)
		}
	}

	return err
}

func (c *grpcClient) Close() {
	c.bg.Close()
}

func init() {
	client.Register(newGRPCClient())
}
