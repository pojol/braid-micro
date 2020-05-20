package dispatcher

import "time"

// Config 调用器配置项
type config struct {
	ConsulAddress string

	PoolInitNum  int
	PoolCapacity int
	PoolIdle     time.Duration

	Tracing bool
}

// Option config wraps
type Option func(*Client)

// WithTracing 开启分布式追踪
func WithTracing() Option {
	return func(r *Client) {
		r.cfg.Tracing = true
	}
}

// WithPoolInitNum 连接池初始化数量
func WithPoolInitNum(num int) Option {
	return func(r *Client) {
		r.cfg.PoolInitNum = num
	}
}

// WithPoolCapacity 连接池的容量大小
func WithPoolCapacity(num int) Option {
	return func(r *Client) {
		r.cfg.PoolCapacity = num
	}
}

// WithPoolIdle 连接池的最大闲置时间
func WithPoolIdle(second int) Option {
	return func(r *Client) {
		r.cfg.PoolIdle = time.Duration(second) * time.Second
	}
}
