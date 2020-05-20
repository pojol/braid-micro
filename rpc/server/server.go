package server

import (
	"errors"
	"io"
	"net"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/tracer"
	"google.golang.org/grpc"
)

type (
	// IServer RPC 服务端
	IServer interface {
		Run()
		Close()
	}

	// Server RPC 服务端
	Server struct {
		rpc          *grpc.Server
		tracerCloser io.Closer

		cfg config
	}
)

var (
	server *Server

	// ErrServiceUnavailiable 没有可用的服务
	ErrServiceUnavailiable = errors.New("service not registered")
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

// New 构建service
func New(name string, opts ...Option) IServer {
	const (
		defaultTracing       = false
		defaultListenAddress = ":14222"
	)

	server = &Server{
		cfg: config{
			Name:          name,
			Tracing:       defaultTracing,
			ListenAddress: defaultListenAddress,
		},
	}

	for _, opt := range opts {
		opt(server)
	}

	if server.cfg.Tracing {
		server.rpc = grpc.NewServer(tracer.GetGRPCServerTracer())
	} else {
		server.rpc = grpc.NewServer()
	}

	return server
}

// Get 获取rpc 服务器
func Get() *grpc.Server {
	return server.rpc
}

// Run 运行
func (s *Server) Run() {

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
func (s *Server) Close() {
	s.rpc.Stop()
}
