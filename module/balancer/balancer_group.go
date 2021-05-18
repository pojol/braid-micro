// balancer 负载均衡模块
//
package balancer

import (
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
)

// IBalancerGroup 负载均衡器
type IBalancerGroup interface {
	module.IModule

	// Pick 为 target 服务选取一个合适的节点
	//
	// strategy 选取所使用的策略，在构建阶段通过 opt 传入
	Pick(strategy string, target string) (discover.Node, error)
}
