package grpcclient

import (
	"time"
)

// Config 调用器配置项
type Config struct {
	Name string

	PoolInitNum  int
	PoolCapacity int
	PoolIdle     time.Duration
}

// Option config wraps
type Option func(*Config)

// WithPoolInitNum 连接池初始化数量
func WithPoolInitNum(num int) Option {
	return func(c *Config) {
		c.PoolInitNum = num
	}
}

// WithPoolCapacity 连接池的容量大小
func WithPoolCapacity(num int) Option {
	return func(c *Config) {
		c.PoolCapacity = num
	}
}

// WithPoolIdle 连接池的最大闲置时间
func WithPoolIdle(second int) Option {
	return func(c *Config) {
		c.PoolIdle = time.Duration(second) * time.Second
	}
}
