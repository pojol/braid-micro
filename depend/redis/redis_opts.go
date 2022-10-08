package redis

import "time"

// Parm 配置项
type Parm struct {
	Address string //connection string, like "redis:// :password@10.0.1.11:6379/0"

	ReadTimeOut    time.Duration
	WriteTimeOut   time.Duration //
	ConnectTimeOut time.Duration //
	MaxIdle        int           //
	MaxActive      int           //
	IdleTimeout    time.Duration
}

// Option config wraps
type Option func(*Parm)

// WithAddr redis 连接地址 "redis:// :password@10.0.1.11:6379/0"
func WithAddr(addr string) Option {
	return func(c *Parm) {
		c.Address = addr
	}
}

// WithReadTimeOut 连接的读取超时时间
func WithReadTimeOut(readtimeout time.Duration) Option {
	return func(c *Parm) {
		c.ReadTimeOut = readtimeout
	}
}

// WithWriteTimeOut 连接的写入超时时间
func WithWriteTimeOut(writetimeout time.Duration) Option {
	return func(c *Parm) {
		c.WriteTimeOut = writetimeout
	}
}

// WithConnectTimeOut 连接超时时间
func WithConnectTimeOut(connecttimeout time.Duration) Option {
	return func(c *Parm) {
		c.ConnectTimeOut = connecttimeout
	}
}

// WithIdleTimeout 闲置连接的超时时间, 设置小于服务器的超时时间 redis.conf : timeout
func WithIdleTimeout(idletimeout time.Duration) Option {
	return func(c *Parm) {
		c.IdleTimeout = idletimeout
	}
}

// WithMaxIdle 最大空闲连接数
func WithMaxIdle(maxidle int) Option {
	return func(c *Parm) {
		c.MaxIdle = maxidle
	}
}

// WithMaxActive 最大连接数，当为0时没有连接数限制
func WithMaxActive(maxactive int) Option {
	return func(c *Parm) {
		c.MaxActive = maxactive
	}
}
