package grpcserver

// Parm Service 配置
type Parm struct {
	ListenAddr string
	isTracing  bool
}

// Option config wraps
type Option func(*Parm)

// WithListen 服务器侦听地址配置
func WithListen(address string) Option {
	return func(c *Parm) {
		c.ListenAddr = address
	}
}

// WithTracing with tracing
func WithTracing() Option {
	return func(c *Parm) {
		c.isTracing = true
	}
}
