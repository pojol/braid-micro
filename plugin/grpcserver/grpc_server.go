package grpcserver

import (
	"errors"
	"io"
	"net"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/module/tracer"
	"google.golang.org/grpc"
)

type grpcServerBuilder struct {
	cfg Config
}

func newGRPCServer() server.Builder {
	return &grpcServerBuilder{}
}

func (b *grpcServerBuilder) Name() string {
	return ServerName
}

func (b *grpcServerBuilder) SetCfg(cfg interface{}) error {
	cecfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	b.cfg = cecfg
	return nil
}

func (b *grpcServerBuilder) Build(tracing bool) server.ISserver {
	s := &grpcServer{
		cfg:     b.cfg,
		tracing: tracing,
	}

	log.Debugf("build grpc server tracing %v", tracing)

	if tracing {
		s.rpc = grpc.NewServer(tracer.GetGRPCServerTracer())
	} else {
		s.rpc = grpc.NewServer()
	}

	return s
}

// Server RPC 服务端
type grpcServer struct {
	rpc          *grpc.Server
	tracing      bool
	tracerCloser io.Closer

	cfg Config
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

	rpcListen, err := net.Listen("tcp", s.cfg.ListenAddress)
	if err != nil {
		log.SysError("register", "run listen "+s.cfg.ListenAddress, err.Error())
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
