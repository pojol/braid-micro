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

// OpenTracing with tracing
func OpenTracing(tracer opentracing.Tracer) Option {
	return func(c *Parm) {
		c.tracer = tracer
	}
}
