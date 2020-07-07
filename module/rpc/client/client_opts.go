package client

import (
	"time"

	"github.com/pojol/braid/plugin/discover"
	"github.com/pojol/braid/plugin/discover/consuldiscover"
)

// Config 调用器配置项
type config struct {
	Name string

	consulCfg consuldiscover.Cfg

	PoolInitNum  int
	PoolCapacity int
	PoolIdle     time.Duration

	Tracing bool
	Link    bool
}

// Option config wraps
type Option func(*Client)

// WithTracing 开启分布式追踪
func WithTracing() Option {
	return func(r *Client) {
		r.cfg.Tracing = true
	}
}

// WithLink 开启链接器
func WithLink() Option {
	return func(r *Client) {
		r.cfg.Link = true
	}
}

// WithConsul 使用consul作为发现器支持
func WithConsul(address string, discoverTag string) Option {
	return func(r *Client) {
		r.cfg.consulCfg.ConsulAddress = address
		r.cfg.consulCfg.Interval = time.Second * 2
		r.cfg.consulCfg.Name = r.cfg.Name
		r.cfg.consulCfg.Tag = discoverTag

		r.discovBuilder = discover.GetBuilder(consuldiscover.DiscoverName)
		err := r.discovBuilder.SetCfg(r.cfg.consulCfg)
		if err != nil {
			// Fatal log
		}
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
