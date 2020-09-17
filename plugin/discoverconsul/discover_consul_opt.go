package discoverconsul

import (
	"time"

	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/pubsub"
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

	procPB    pubsub.IPubsub
	clusterPB pubsub.IPubsub
	linkcache linkcache.ILinkCache
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

// WithConsulAddress with consul address
func WithConsulAddress(address string) Option {
	return func(c *Parm) {
		c.Address = address
	}
}

// WithProcPubsub with proc
func WithProcPubsub(pb pubsub.IPubsub) Option {
	return func(c *Parm) {
		c.procPB = pb
	}
}

// WithClusterPubsub with cluster
func WithClusterPubsub(pb pubsub.IPubsub) Option {
	return func(c *Parm) {
		c.clusterPB = pb
	}
}

// WithLinkCache with link cache
func WithLinkCache(cache linkcache.ILinkCache) Option {
	return func(c *Parm) {
		c.linkcache = cache
	}
}
