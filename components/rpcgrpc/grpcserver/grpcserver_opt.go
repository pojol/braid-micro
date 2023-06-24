package grpcserver

import (
	"google.golang.org/grpc"
)

type RegistHandler func(*grpc.Server)

// Parm Service 配置
type Parm struct {
	ListenAddr string

	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor

	Handler RegistHandler

	GracefulStop bool
}

// Option config wraps
type Option func(*Parm)

// WithListen 服务器侦听地址配置
func WithListen(address string) Option {
	return func(c *Parm) {
		c.ListenAddr = address
	}
}

func WithGracefulStop() Option {
	return func(c *Parm) {
		c.GracefulStop = true
	}
}

func AppendUnaryInterceptors(interceptor grpc.UnaryServerInterceptor) Option {
	return func(c *Parm) {
		c.UnaryInterceptors = append(c.UnaryInterceptors, interceptor)
	}
}

func AppendStreamInterceptors(interceptor grpc.StreamServerInterceptor) Option {
	return func(c *Parm) {
		c.StreamInterceptors = append(c.StreamInterceptors, interceptor)
	}
}

func RegisterHandler(handler RegistHandler) Option {
	return func(c *Parm) {
		c.Handler = handler
	}
}
