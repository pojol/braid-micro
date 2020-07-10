package grpcserver

// Config Service 配置
type Config struct {
	Tracing       bool
	Name          string
	ListenAddress string
}

// Option config wraps
type Option func(*Config)

// WithTracing 开启分布式追踪
func WithTracing() Option {
	return func(c *Config) {
		c.Tracing = true
	}
}

// WithListen 服务器侦听地址配置
func WithListen(address string) Option {
	return func(c *Config) {
		c.ListenAddress = address
	}
}
