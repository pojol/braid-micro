package linkerredis

import (
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/pubsub"
)

// Parm link-cache with redis parm
type Parm struct {
	ServiceName string
	elector     elector.IElection
	clusterPB   pubsub.IPubsub
}

// Option consul discover config wrapper
type Option func(*Parm)

// WithElector with elector
func WithElector(e elector.IElection) Option {
	return func(c *Parm) {
		c.elector = e
	}
}

// WithClusterPubsub with cluster
func WithClusterPubsub(pb pubsub.IPubsub) Option {
	return func(c *Parm) {
		c.clusterPB = pb
	}
}
