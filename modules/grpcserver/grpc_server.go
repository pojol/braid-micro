package grpcserver

import (
	"errors"
	"net"

	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/rpc/server"
	"github.com/pojol/braid-go/modules/jaegertracing"
	"google.golang.org/grpc"
)

var (
	// Name grpc plugin name
	Name = "GRPCServer"

	// ErrServiceUnavailiable 没有可用的服务
	ErrServiceUnavailiable = errors.New("service not registered")
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
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
	return Name
}

func (b *grpcServerBuilder) Build(serviceName string, logger logger.ILogger) (server.IServer, error) {
	p := Parm{
		ListenAddr: ":14222",
	}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	s := &grpcServer{
		parm:        p,
		logger:      logger,
		serviceName: serviceName,
	}

	var istracing bool
	if p.tracer != nil {
		istracing = true
		s.rpc = grpc.NewServer(jaegertracing.GetGRPCServerTracer(p.tracer))
	} else {
		s.rpc = grpc.NewServer()
	}

	s.logger.Debugf("build grpc-server listen: %s tracing: %t", p.ListenAddr, istracing)
	return s, nil
}

// Server RPC 服务端
type grpcServer struct {
	rpc         *grpc.Server
	serviceName string

	logger logger.ILogger
	parm   Parm
}

func (s *grpcServer) Init() {

}

// Get 获取rpc 服务器
func (s *grpcServer) Server() interface{} {
	return s.rpc
}

// Run 运行
func (s *grpcServer) Run() {

	rpcListen, err := net.Listen("tcp", s.parm.ListenAddr)
	if err != nil {
		s.logger.Errorf("server listen err %s %s", err.Error(), s.parm.ListenAddr)
	}

	go func() {
		if err := s.rpc.Serve(rpcListen); err != nil {
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
	server.Register(newGRPCServer())
}
