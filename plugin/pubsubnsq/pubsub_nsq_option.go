package pubsubnsq

import "github.com/nsqio/go-nsq"

// NsqConfig nsq config
type NsqConfig struct {
	nsqCfg *nsq.Config

	LookupAddres []string
	Addres       []string

	Channel string
}

// Option config wraps
type Option func(*NsqConfig)

// WithChannel 通过channel 构建
func WithChannel(channel string) Option {
	return func(c *NsqConfig) {
		c.Channel = channel
	}
}
