package grpcserver

import (
	"errors"
	"net"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
	"google.golang.org/grpc"
)

type grpcServerBuilder struct {
	opts []interface{}
}

func newGRPCServer() server.Builder {
	return &grpcServerBuilder{}
}

func (b *grpcServerBuilder) AddOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *grpcServerBuilder) Name() string {
	return ServerName
}

func (b *grpcServerBuilder) Build(serviceName string) (server.ISserver, error) {
	p := Parm{
		ListenAddr: ":14222",
	}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	s := &grpcServer{
		parm:        p,
		serviceName: serviceName,
	}

	if p.isTracing {
		s.rpc = grpc.NewServer(tracer.GetGRPCServerTracer())
	} else {
		s.rpc = grpc.NewServer()
	}

	return s, nil
}

// Server RPC 服务端
type grpcServer struct {
	rpc         *grpc.Server
	serviceName string

	parm Parm
}

var (
	// ServerName grpc plugin name
	ServerName = "GRPCServer"

	// ErrServiceUnavailiable 没有可用的服务
	ErrServiceUnavailiable = errors.New("service not registered")
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

// Get 获取rpc 服务器
func (s *grpcServer) Server() interface{} {
	return s.rpc
}

// Run 运行
func (s *grpcServer) Run() {

	rpcListen, err := net.Listen("tcp", s.parm.ListenAddr)
	if err != nil {
		log.SysError("register", "run listen "+s.parm.ListenAddr, err.Error())
	}

	go func() {
		if err := s.rpc.Serve(rpcListen); err != nil {
			log.SysError("register", "run serve", err.Error())
		}
	}()
}

// Close 退出处理
func (s *grpcServer) Close() {
	s.rpc.Stop()
}

func init() {
	server.Register(newGRPCServer())
}
