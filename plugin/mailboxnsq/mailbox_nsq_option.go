package mailboxnsq

import "github.com/nsqio/go-nsq"

// Parm nsq config
type Parm struct {
	nsqCfg nsq.Config

	LookupAddres []string
	Addres       []string

	Ephemeral bool

	Channel     string
	ServiceName string
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
		c.LookupAddres = addr
	}
}

// WithNsqdAddr nsqd addr
func WithNsqdAddr(addr []string) Option {
	return func(c *Parm) {
		c.Addres = addr
	}
}

// WithEphemeral open ephemeral
func WithEphemeral() Option {
	return func(c *Parm) {
		c.Ephemeral = true
	}
}
