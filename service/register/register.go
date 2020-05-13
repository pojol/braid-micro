package register

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/service/dispatcher/bproto"
	"github.com/pojol/braid/tracer"
	"google.golang.org/grpc"
)

type (
	// Register 注册器
	Register struct {
		rpc          *grpc.Server
		tracerCloser io.Closer
		listen       string
	}

	// Config Service 配置
	Config struct {
		Tracing       bool
		Name          string
		ListenAddress string
	}

	// RPCFunc ...
	// ctx 上下文
	// in 外部发送过来的数据报文
	// out 返回给外部的数据报文
	// err 错误信息
	RPCFunc func(ctx context.Context, in []byte) (out []byte, err error)
)

var (
	serviceMap map[string]RPCFunc = make(map[string]RPCFunc)

	register *Register

	// DefaultConfig 默认配置
	DefaultConfig = Config{
		Tracing:       false,
		Name:          "defaultRegistName",
		ListenAddress: ":14222",
	}

	// ErrServiceUnavailiable 没有可用的服务
	ErrServiceUnavailiable = errors.New("service unavailable")
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

// New 构建service
func New() *Register {
	register = &Register{}
	return register
}

// Init 构建service
func (s *Register) Init(cfg interface{}) error {

	sCfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	var rpcServer *grpc.Server
	var err error
	var closer io.Closer

	if sCfg.Tracing {
		rpcServer = grpc.NewServer(tracer.GetGRPCServerTracer())
	} else {
		rpcServer = grpc.NewServer()
	}

	s.rpc = rpcServer
	s.tracerCloser = closer
	s.listen = sCfg.ListenAddress

	return err
}

type rpcServer struct {
	bproto.ListenServer
}

func (s *rpcServer) Routing(ctx context.Context, in *bproto.RouteReq) (*bproto.RouteRes, error) {

	var err error
	var body []byte

	if _, ok := serviceMap[in.Service]; !ok {
		err = ErrServiceUnavailiable
		goto EXT
	}

	body, err = serviceMap[in.Service](ctx, in.GetReqBody())
	if err != nil {
		goto EXT
	}

EXT:
	if err != nil {
		log.SysError("main", "routing", err.Error())
	}

	return &bproto.RouteRes{ResBody: body}, err
}

// Regist 注册服务
func (s *Register) Regist(serviceName string, fc RPCFunc) {
	if _, ok := serviceMap[serviceName]; ok {
		return
	}

	serviceMap[serviceName] = fc
}

// Run 运行
func (s *Register) Run() {
	bproto.RegisterListenServer(s.rpc, &rpcServer{})

	rpcListen, err := net.Listen("tcp", s.listen)
	if err != nil {
		log.SysError("register", "run listen", err.Error())
	}

	go func() {
		if err := s.rpc.Serve(rpcListen); err != nil {
			log.SysError("register", "run serve", err.Error())
		}
	}()
}

// Close 退出处理
func (s *Register) Close() {
	s.rpc.Stop()
}
