package balancer

import (
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/discover"
)

// IBalancer 负载均衡
type IBalancer interface {
	module.IModule

	// 从服务节点列表中选取一个对应的节点，
	// 节点列表可以订阅discover模块的消息进行填充或更改，
	// braid 提供默认的`平滑加权轮询算法`如果有其他的需求，用户可以选择实现自定义的Pick接口。
	Pick() (nod discover.Node, err error)

	// 随机选取一个节点（
	Random() (nod discover.Node, err error)
}
