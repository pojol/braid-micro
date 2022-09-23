package consul

import (
	"crypto/tls"
	"time"

	consul "github.com/hashicorp/consul/api"
)

type Parm struct {
	cfg *consul.Config

	connect bool

	queryOpt *consul.QueryOptions

	allowStale bool // 主从节点都可以进行数据读取

	Addrs []string

	timeout time.Duration

	tlsConfig *tls.Config
}

type Option func(*Parm)

func WithConfig(cfg *consul.Config) Option {
	return func(c *Parm) {
		c.cfg = cfg
	}
}

func WithQueryOption(opt *consul.QueryOptions) Option {
	return func(c *Parm) {
		c.queryOpt = opt
	}
}

func WithAllowStale(as bool) Option {
	return func(c *Parm) {
		c.allowStale = as
	}
}

func WithAddress(address []string) Option {
	return func(c *Parm) {
		c.Addrs = address
	}
}

func WithTimeOut(timeout time.Duration) Option {
	return func(c *Parm) {
		c.timeout = timeout
	}
}

func WithTLS(tls *tls.Config) Option {
	return func(c *Parm) {
		c.tlsConfig = tls
	}
}
