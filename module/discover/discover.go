package discover

import (
	"encoding/json"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/mailbox"
)

const (
	ServiceUpdate = "discover.serviceUpdate"

	EventAddService    = "event_add_service"
	EventRemoveService = "event_remove_service"
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

func EncodeUpdateMsg(event string, nod Node) *mailbox.Message {
	byt, _ := json.Marshal(&UpdateMsg{
		Event: event,
		Nod:   nod,
	})

	return &mailbox.Message{
		Body: byt,
	}
}

func DecodeUpdateMsg(msg *mailbox.Message) UpdateMsg {
	dmsg := UpdateMsg{}
	json.Unmarshal(msg.Body, &dmsg)
	return dmsg
}

// IDiscover discover interface
type IDiscover interface {
	module.IModule
}
