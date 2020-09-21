package electork8s

import "time"

// Parm k8s elector config
type Parm struct {
	ServiceName string
	KubeCfg     string
	NodID       string
	Namespace   string
	RetryPeriod time.Duration
}

// Option consul discover config wrapper
type Option func(*Parm)

// WithKubeConfig with kube config
func WithKubeConfig(config string) Option {

	return func(c *Parm) {
		c.KubeCfg = config
	}

}

// WithNodID with nod id
func WithNodID(nodid string) Option {
	return func(c *Parm) {
		c.NodID = nodid
	}
}

// WithNamespace with name space
func WithNamespace(namespace string) Option {
	return func(c *Parm) {
		c.Namespace = namespace
	}
}

// WithRetryTick with retry tick
func WithRetryTick(t time.Duration) Option {
	return func(c *Parm) {
		c.RetryPeriod = t
	}
}
