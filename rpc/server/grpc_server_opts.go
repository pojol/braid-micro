package server

import (
	"google.golang.org/grpc"
)

// Parm Service 配置
type Parm struct {
	ListenAddr string

	interceptors []grpc.UnaryServerInterceptor
}

// Option config wraps
type Option func(*Parm)

// WithListen 服务器侦听地址配置
func WithListen(address string) Option {
	return func(c *Parm) {
		c.ListenAddr = address
	}
}

func AppendInterceptors(interceptor grpc.UnaryServerInterceptor) Option {
	return func(c *Parm) {
		c.interceptors = append(c.interceptors, interceptor)
	}
}
