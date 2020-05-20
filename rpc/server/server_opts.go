package server

// Config Service 配置
type config struct {
	Tracing       bool
	Name          string
	ListenAddress string
}

// Option config wraps
type Option func(*Server)

// WithTracing 开启分布式追踪
func WithTracing() Option {
	return func(r *Server) {
		r.cfg.Tracing = true
	}
}

// WithListen 服务器侦听地址配置
func WithListen(address string) Option {
	return func(r *Server) {
		r.cfg.ListenAddress = address
	}
}
