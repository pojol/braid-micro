// 实现文件 grpcserver 基于 grpc 实现的 rpc-server
package server

import (
	"errors"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/tracer"
	"github.com/pojol/braid-go/module/rpc/server"
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

func BuildWithOption(serviceName string, opts ...server.Option) server.IServer {

	p := server.Parm{
		ListenAddr: ":14222",
	}
	for _, opt := range opts {
		opt(&p)
	}

	s := &grpcServer{
		parm:        p,
		serviceName: serviceName,
	}

	if p.Tracer != nil {
		s.tracer = p.Tracer.GetTracing().(opentracing.Tracer)
		p.Interceptors = append(p.Interceptors, tracer.ServerInterceptor(s.tracer))
	}

	if len(p.Interceptors) != 0 {
		s.rpc = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(p.Interceptors...)))
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
	parm   server.Parm

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
			blog.Errf("run server err %s", err.Error())
		}
	}()
}

// Close 退出处理
func (s *grpcServer) Close() {
	blog.Debugf("grpc-server closed")
	s.rpc.Stop()
}
