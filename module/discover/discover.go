package discover

import (
	"encoding/json"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/mailbox"
)

// discover
const (
	// DiscoverAddService publish topic
	DiscoverAddService = "topic_discover_add_service"
	// DiscoverRmvService publish topic
	DiscoverRmvService = "topic_discover_remove_service"
	// DiscoverUpdateService publish topic
	DiscoverUpdateService = "topic_discover_update_service"
)

// RmvServiceMsg down msg
type RmvServiceMsg struct {
	ID      string
	Service string
	Addr    string
}

// EncodeDownMsg encode down msg
func EncodeRmvServiceMsg(id string, service string, addr string) *mailbox.Message {
	byt, _ := json.Marshal(&RmvServiceMsg{
		ID:      id,
		Service: service,
		Addr:    addr,
	})

	return &mailbox.Message{
		Body: byt,
	}
}

// DecodeDownMsg decode down msg
func DecodeRmvServiceMsg(msg *mailbox.Message) RmvServiceMsg {
	dmsg := RmvServiceMsg{}
	json.Unmarshal(msg.Body, &dmsg)
	return dmsg
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

// IDiscover discover interface
type IDiscover interface {
	module.IModule
}
