package grpcserver

import "github.com/opentracing/opentracing-go"

// Parm Service 配置
type Parm struct {
	ListenAddr string
	tracer     opentracing.Tracer
}

// Option config wraps
type Option func(*Parm)

// WithListen 服务器侦听地址配置
func WithListen(address string) Option {
	return func(c *Parm) {
		c.ListenAddr = address
	}
}

// AutoOpenTracing 打开tracing
//
// 当 tracing 被注册到braid中后，braid在构建过程中会自动引用这个函数，将tracer自动绑定到server模块
func AutoOpenTracing(tracer opentracing.Tracer) Option {
	return func(c *Parm) {
		c.tracer = tracer
	}
}
