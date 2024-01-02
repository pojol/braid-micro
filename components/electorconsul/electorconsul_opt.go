package electorconsul

import (
	"time"

	"github.com/pojol/braid-go/components/depends/bconsul"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
)

// Parm 选举器配置项
type Parm struct {
	LockTick          time.Duration
	RefushSessionTick time.Duration

	ConsulCli *bconsul.Client

	Pubsub module.IPubsub

	Log *blog.Logger
}

// Option consul discover config wrapper
type Option func(*Parm)

// WithLockTick with lock tick
func WithLockTick(t time.Duration) Option {
	return func(c *Parm) {
		c.LockTick = t
	}
}

// WithSessionTick with session tick
func WithSessionTick(t time.Duration) Option {
	return func(c *Parm) {
		c.RefushSessionTick = t
	}
}

func WithLog(log *blog.Logger) Option {
	return func(c *Parm) {
		c.Log = log
	}
}

func WithPubsub(pubsub module.IPubsub) Option {
	return func(c *Parm) {
		c.Pubsub = pubsub
	}
}

func WithConsulClient(cli *bconsul.Client) Option {
	return func(c *Parm) {
		c.ConsulCli = cli
	}
}
