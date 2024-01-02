package discoverconsul

import (
	"time"
)

// Parm discover config
type Parm struct {

	// 同步节点信息间隔
	SyncServicesInterval time.Duration

	// 同步节点权重间隔
	SyncServiceWeightInterval time.Duration

	// 过滤标签
	Tag string

	// 黑名单列表，在黑名单的服务将不会被加入到 services 中
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
