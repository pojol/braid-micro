package grpcclient

import (
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pojol/braid/module/linkcache"
)

// Parm 调用器配置项
type Parm struct {
	Name string

	tracer opentracing.Tracer

	byLink bool
	linker linkcache.ILinkCache

	balancerStrategy []string
	balancerGroup    string

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

// AutoOpenTracing 打开tracing
//
// 当 tracing 被注册到braid中后，braid在构建过程中会自动引用这个函数，将tracer自动绑定到client模块
func AutoOpenTracing(tracer opentracing.Tracer) Option {
	return func(c *Parm) {
		c.tracer = tracer
	}
}

// AutoLinkCache 绑定链路缓存
//
// 当linkcache 被注册到braid中后，braid在构建过程中会自动引用这个函数，将linkcache自动绑定到client模块
func AutoLinkCache(cache linkcache.ILinkCache) Option {
	return func(c *Parm) {
		c.byLink = true
		c.linker = cache
	}
}

// WithBalanceStrategy 添加负载均衡选用的策略
func WithBalanceStrategy(strategies []string) Option {
	return func(c *Parm) {
		c.balancerStrategy = strategies
	}
}

// WithBalanceGroup 挂载到client的负载均衡控制器
func WithBalanceGroup(bg string) Option {
	return func(c *Parm) {
		c.balancerGroup = bg
	}
}
