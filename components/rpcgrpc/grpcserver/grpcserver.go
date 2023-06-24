// 实现文件 grpcserver 基于 grpc 实现的 rpc-server
package grpcserver

import (
	"errors"
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
	"google.golang.org/grpc"
)

var (
	// ErrServiceUnavailiable 没有可用的服务
	ErrServiceUnavailiable = errors.New("service not registered")
)

// Server RPC 服务端
type grpcServer struct {
	rpc  *grpc.Server
	info meta.ServiceInfo

	listen net.Listener
	log    *blog.Logger
	parm   Parm
}

func BuildWithOption(info meta.ServiceInfo, log *blog.Logger, opts ...Option) module.IServer {

	p := Parm{
		ListenAddr: ":14222",
	}

	for _, opt := range opts {
		opt(&p)
	}

	var rpcserver *grpc.Server

	if len(p.UnaryInterceptors) != 0 {
		rpcserver = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(p.UnaryInterceptors...)))
	} else {
		rpcserver = grpc.NewServer()
	}

	if p.Handler == nil {
		panic(fmt.Errorf("grpc server handler not set"))
	}

	return &grpcServer{
		info: info,
		parm: p,
		log:  log,
		rpc:  rpcserver,
	}

}

func (s *grpcServer) Init() error {

	rpcListen, err := net.Listen("tcp", s.parm.ListenAddr)
	if err != nil {
		return fmt.Errorf("%v [GRPC] server check error %v [%v]", s.info.Name, "tcp", s.parm.ListenAddr)
	} else {
		s.log.Infof("[GRPC] server listen: [tcp] %v", s.parm.ListenAddr)
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

	// regist rpc handler
	s.parm.Handler(s.rpc)

	go func() {
		if err := s.rpc.Serve(s.listen); err != nil {
			s.log.Errf("[GRPC] server serving err %s", err.Error())
		}
	}()

}

// Close 退出处理
func (s *grpcServer) Close() {
	s.log.Infof("grpc-server closed")
	if s.parm.GracefulStop {
		s.rpc.GracefulStop()
	} else {
		s.rpc.Stop()
	}
}
