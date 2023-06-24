package grpcclient

import (
	"time"

	"google.golang.org/grpc"
)

// Parm 调用器配置项
type Parm struct {
	PoolInitNum  int
	PoolCapacity int
	PoolIdle     time.Duration

	UnaryInterceptors  []grpc.UnaryClientInterceptor
	StreamInterceptors []grpc.StreamClientInterceptor
}

var (
	DefaultClientParm = Parm{
		PoolInitNum:  8,
		PoolCapacity: 64,
		PoolIdle:     time.Second * 100,
	}
)

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

func AppendUnaryInterceptors(interceptor grpc.UnaryClientInterceptor) Option {
	return func(c *Parm) {
		c.UnaryInterceptors = append(c.UnaryInterceptors, interceptor)
	}
}

func AppendStreamInterceptors(interceptor grpc.StreamClientInterceptor) Option {
	return func(c *Parm) {
		c.StreamInterceptors = append(c.StreamInterceptors, interceptor)
	}
}
