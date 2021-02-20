package balancer

import (
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
)

// IBalancerGroup balancer group interface
type IBalancerGroup interface {
	module.IModule

	Pick(ty string, target string) (discover.Node, error)
}
