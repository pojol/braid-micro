package discoverconsul

import "time"

// Cfg discover config
type Cfg struct {
	Name string

	// 同步节点信息间隔
	Interval time.Duration

	// 注册中心
	Address string

	Tag string

	Blacklist []string
}

// Option consul discover config wrapper
type Option func(*Cfg)

// WithTag 修改config中的discover tag
func WithTag(discoverTag string) Option {
	return func(c *Cfg) {
		c.Tag = discoverTag
	}
}

// WithBlacklist add blacklist
func WithBlacklist(lst []string) Option {
	return func(c *Cfg) {
		c.Blacklist = lst
	}
}

// WithInterval 修改config中的interval
func WithInterval(interval time.Duration) Option {
	return func(c *Cfg) {
		c.Interval = interval
	}
}
