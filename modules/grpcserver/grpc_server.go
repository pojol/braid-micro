// 实现文件 grpcserver 基于 grpc 实现的 rpc-server
package grpcserver

import (
	"errors"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/modules/jaegertracing"
	"github.com/pojol/braid-go/modules/moduleparm"
	"google.golang.org/grpc"
)

var (
	// Name grpc plugin name
	Name = "GRPCServer"

	// ErrServiceUnavailiable 没有可用的服务
	ErrServiceUnavailiable = errors.New("service not registered")
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert linker config")
)

type grpcServerBuilder struct {
	opts []interface{}
}

func newGRPCServer() module.IBuilder {
	return &grpcServerBuilder{}
}

func (b *grpcServerBuilder) AddModuleOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *grpcServerBuilder) Name() string {
	return Name
}

func (b *grpcServerBuilder) Type() module.ModuleType {
	return module.Server
}

func (b *grpcServerBuilder) Build(serviceName string, buildOpts ...interface{}) interface{} {
	bp := moduleparm.BuildParm{}
	for _, opt := range buildOpts {
		opt.(moduleparm.Option)(&bp)
	}

	p := Parm{
		ListenAddr:  ":14222",
		openRecover: false,
	}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	s := &grpcServer{
		parm:        p,
		logger:      bp.Logger,
		serviceName: serviceName,
	}

	interceptors := []grpc.UnaryServerInterceptor{}
	if bp.Tracer != nil {
		s.tracer = bp.Tracer.GetTracing().(opentracing.Tracer)
		interceptors = append(interceptors, jaegertracing.ServerInterceptor(s.tracer))
	}

	if p.openRecover {
		interceptors = append(interceptors,
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(p.recoverHandle)))
	}

	if len(interceptors) != 0 {
		s.rpc = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)))
	} else {
		s.rpc = grpc.NewServer()
	}

	return s
}

// Server RPC 服务端
type grpcServer struct {
	rpc         *grpc.Server
	serviceName string

	listen net.Listener
	logger logger.ILogger
	parm   Parm

	tracer opentracing.Tracer
}

func (s *grpcServer) Init() error {

	rpcListen, err := net.Listen("tcp", s.parm.ListenAddr)
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", s.serviceName, "tcp", s.parm.ListenAddr)
	}

	s.listen = rpcListen

	return nil
}

// Get 获取rpc 服务器
func (s *grpcServer) Server() interface{} {
	return s.rpc
}

// Run 运行
func (s *grpcServer) Run() {

	go func() {
		if err := s.rpc.Serve(s.listen); err != nil {
			s.logger.Errorf("run server err %s", err.Error())
		}
	}()
}

// Close 退出处理
func (s *grpcServer) Close() {
	s.logger.Debugf("grpc-server closed")
	s.rpc.Stop()
}

func init() {
	module.Register(newGRPCServer())
}
