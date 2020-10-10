package grpcclient

import (
	"time"

	"github.com/pojol/braid/module/linkcache"
)

// Parm 调用器配置项
type Parm struct {
	Name string

	isTracing bool

	byLink bool
	linker linkcache.ILinkCache

	PoolInitNum  int
	PoolCapacity int
	PoolIdle     time.Duration
}

// Option config wraps
type Option func(*Parm)

// WithPoolInitNum 连接池初始化数量
func WithPoolInitNum(num int) Option {
	return func(c *Parm) {
		c.PoolInitNum = num
	}
}

// WithPoolCapacity 连接池的容量大小
func WithPoolCapacity(num int) Option {
	return func(c *Parm) {
		c.PoolCapacity = num
	}
}

// WithPoolIdle 连接池的最大闲置时间
func WithPoolIdle(second int) Option {
	return func(c *Parm) {
		c.PoolIdle = time.Duration(second) * time.Second
	}
}

// Tracing open tracing (auto register)
func Tracing() Option {
	return func(c *Parm) {
		c.isTracing = true
	}
}

// LinkCache with link-cache (auto register)
func LinkCache(cache linkcache.ILinkCache) Option {
	return func(c *Parm) {
		c.byLink = true
		c.linker = cache
	}
}
