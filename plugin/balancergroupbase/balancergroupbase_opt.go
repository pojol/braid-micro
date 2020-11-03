package balancergroupbase

// Parm balancer group parm
type Parm struct {
	strategies []string
}

// Option parm opt
type Option func(*Parm)

// WithStrategy add strategy
func WithStrategy(strategies []string) Option {
	return func(c *Parm) {
		c.strategies = strategies
	}
}
