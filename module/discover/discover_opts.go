package discover

import (
	"time"

	"github.com/pojol/braid-go/depend/consul"
)

// Parm discover config
type Parm struct {
	Name string

	// 同步节点信息间隔
	SyncServicesInterval time.Duration

	// 同步节点权重间隔
	SyncServiceWeightInterval time.Duration

	// 注册中心
	Address string

	Tag string

	Blacklist []string

	Client *consul.Client
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

// WithSyncServiceInterval 修改config中的interval
func WithSyncServiceInterval(interval time.Duration) Option {
	return func(c *Parm) {
		c.SyncServicesInterval = interval
	}
}

// WithSyncServiceWeightInterval 修改权重同步间隔
func WithSyncServiceWeightInterval(interval time.Duration) Option {
	return func(c *Parm) {
		c.SyncServiceWeightInterval = interval
	}
}

func WithConsulClient(c *consul.Client) Option {
	return func(p *Parm) {
		p.Client = c
	}
}
