package grpcserver

// Config Service 配置
type Config struct {
	Name          string
	ListenAddress string
}

// Option config wraps
type Option func(*Config)

// WithListen 服务器侦听地址配置
func WithListen(address string) Option {
	return func(c *Config) {
		c.ListenAddress = address
	}
}
