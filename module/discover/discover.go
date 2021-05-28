// 接口文件 discover 服务发现
//
// 这个模块会创建 ServiceUpdate Topic，通过这个 Topic 发布集群中相关服务的变更信息
package discover

import (
	"encoding/json"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/pubsub"
)

const (
	ServiceUpdate = "discover.serviceUpdate"

	// EventAddService 有一个新的服务加入到集群
	EventAddService = "event_add_service"

	// EventRemoveService 有一个旧的服务从集群中退出
	EventRemoveService = "event_remove_service"

	// EventUpdateService 有一个旧的服务产生了信息的变更（通常是指权重
	EventUpdateService = "event_update_service"
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

type UpdateMsg struct {
	Nod   Node
	Event string
}

func EncodeUpdateMsg(event string, nod Node) *pubsub.Message {
	byt, _ := json.Marshal(&UpdateMsg{
		Event: event,
		Nod:   nod,
	})

	return &pubsub.Message{
		Body: byt,
	}
}

func DecodeUpdateMsg(msg *pubsub.Message) UpdateMsg {
	dmsg := UpdateMsg{}
	json.Unmarshal(msg.Body, &dmsg)
	return dmsg
}

// IDiscover discover interface
type IDiscover interface {
	module.IModule
}
