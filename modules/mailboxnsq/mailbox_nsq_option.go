package mailboxnsq

import "github.com/nsqio/go-nsq"

// Parm nsq config
type Parm struct {
	nsqCfg nsq.Config

	LookupdAddress  []string
	NsqdAddress     []string
	NsqdHttpAddress []string

	Channel     string
	ServiceName string

	ConcurrentHandler int32 // consumer 接收句柄的并发数（默认1

	nsqLogLv nsq.LogLevel
}

// Option config wraps
type Option func(*Parm)

// WithChannel 通过channel 构建
func WithChannel(channel string) Option {
	return func(c *Parm) {
		c.Channel = channel
	}
}

// WithNsqConfig nsq config
func WithNsqConfig(cfg nsq.Config) Option {
	return func(c *Parm) {
		c.nsqCfg = cfg
	}
}

// WithLookupAddr lookup addr
func WithLookupAddr(addr []string) Option {
	return func(c *Parm) {
		c.LookupdAddress = addr
	}
}

// WithNsqdAddr nsqd addr
func WithNsqdAddr(addr []string) Option {
	return func(c *Parm) {
		c.NsqdAddress = addr
	}
}

func WithNsqdHTTPAddr(addr []string) Option {
	return func(c *Parm) {
		c.NsqdHttpAddress = addr
	}
}

// WithNsqLogLv 修改nsq的日志等级
func WithNsqLogLv(lv nsq.LogLevel) Option {
	return func(c *Parm) {
		c.nsqLogLv = lv
	}
}

// WithHandlerConcurrent 消费者接收句柄的并发数量（默认1
func WithHandlerConcurrent(cnt int32) Option {
	return func(c *Parm) {
		c.ConcurrentHandler = cnt
	}
}
