package discover

import (
	"strings"

	"github.com/pojol/braid/module/pubsub"
)

// Builder 构建器接口
type Builder interface {
	Build(ps pubsub.IPubsub) IDiscover
	Name() string
	SetCfg(cfg interface{}) error
}

// Node 发现节点结构
type Node struct {
	ID string
	// 负载均衡节点的名称，这个名称主要用于均衡节点分组。
	Name    string
	Address string

	// 节点的权重值
	Weight int
}

// event
const (
	EventAdd    = "discover_event_add"
	EventRmv    = "discover_event_remove"
	EventUpdate = "discover_event_update"
)

// IDiscover 发现服务 & 注册节点
type IDiscover interface {
	// 实现发现服务
	Discover()

	// 关闭发现服务
	Close()
}

var (
	m = make(map[string]Builder)

	discov IDiscover
)

// Register 注册balancer
func Register(b Builder) {
	m[strings.ToLower(b.Name())] = b
}

// GetBuilder 获取构建器
func GetBuilder(name string) Builder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}
