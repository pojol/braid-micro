package balancerswrr

import "github.com/pojol/braid/module/pubsub"

// Parm discover config
type Parm struct {
	Name string

	procPB pubsub.IPubsub
}

// Option consul discover config wrapper
type Option func(*Parm)

// WithProcPubsub with proc
func WithProcPubsub(pb pubsub.IPubsub) Option {
	return func(c *Parm) {
		c.procPB = pb
	}
}
