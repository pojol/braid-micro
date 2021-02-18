package grpcserver

import (
	"errors"
	"fmt"
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

	if p.tracer != nil {
		s.rpc = grpc.NewServer(jaegertracing.GetGRPCServerTracer(p.tracer))
	} else {
		s.rpc = grpc.NewServer()
	}

	return s, nil
}

// Server RPC 服务端
type grpcServer struct {
	rpc         *grpc.Server
	serviceName string

	listen net.Listener
	logger logger.ILogger
	parm   Parm
}

func (s *grpcServer) Init() error {

	rpcListen, err := net.Listen("tcp", s.parm.ListenAddr)
	if err != nil {
		return fmt.Errorf("Dependency check error %v [%v]", "tcp", s.parm.ListenAddr)
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
	server.Register(newGRPCServer())
}
