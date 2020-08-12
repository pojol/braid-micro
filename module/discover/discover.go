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

// IDiscover 服务发现
type IDiscover interface {
	// 通过Tag发现集群内的服务，
	// 将新的服务信息通知到已订阅的各个模块中，
	// 在braid中主要提供从两种中心里实现发现逻辑（docker部署采用consul，k8s部署采用它提供的接口
	//
	// 注：braid本身没有提供服务注册的接口，在采用docker部署时，
	// 注册由容器注册服务提供（https://github.com/gliderlabs/registrator
	Discover()

	// 关闭服务发现
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
