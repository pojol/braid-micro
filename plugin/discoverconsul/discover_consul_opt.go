package discoverconsul

import (
	"time"
)

// Parm discover config
type Parm struct {
	Name string

	// 同步节点信息间隔
	Interval time.Duration

	// 注册中心
	Address string

	Tag string

	Blacklist []string
}

// Option consul discover config wrapper
type Option func(*Parm)

// WithTag 修改config中的discover tag
func WithTag(discoverTag string) Option {
	return func(c *Parm) {
		c.Tag = discoverTag
	}
}

// WithBlacklist add blacklist
func WithBlacklist(lst []string) Option {
	return func(c *Parm) {
		c.Blacklist = lst
	}
}

// WithInterval 修改config中的interval
func WithInterval(interval time.Duration) Option {
	return func(c *Parm) {
		c.Interval = interval
	}
}

// WithConsulAddr with consul address
func WithConsulAddr(address string) Option {
	return func(c *Parm) {
		c.Address = address
	}
}
