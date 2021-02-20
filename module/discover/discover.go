package discover

import "github.com/pojol/braid-go/module"

// discover
const (
	AddService    = "topic_add_service"
	RmvService    = "topic_rmv_service"
	UpdateService = "topic_update_service"
)

// Node 发现节点结构
type Node struct {
	ID string
	// 负载均衡节点的名称，这个名称主要用于均衡节点分组。
	Name    string
	Address string

	// 节点的权重值
	Weight int
}

// IDiscover discover interface
type IDiscover interface {
	module.IModule
}
