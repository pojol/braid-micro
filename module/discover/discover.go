package discover

import (
	"encoding/json"

	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/service"
)

const (
	TopicServiceUpdate = "discover.serviceUpdate"
)

type DiscoverUpdateMsg struct {
	Nod   service.Node
	Event string
}

func DiscoverEncodeUpdateMsg(event string, nod service.Node) *pubsub.Message {
	byt, _ := json.Marshal(&DiscoverUpdateMsg{
		Event: event,
		Nod:   nod,
	})

	return &pubsub.Message{
		Body: byt,
	}
}

func DiscoverDecodeUpdateMsg(msg *pubsub.Message) DiscoverUpdateMsg {
	dmsg := DiscoverUpdateMsg{}
	json.Unmarshal(msg.Body, &dmsg)
	return dmsg
}

const (
	// EventAddService 有一个新的服务加入到集群
	EventAddService = "event.service.nodeAdd"

	// EventRemoveService 有一个旧的服务从集群中退出
	EventRemoveService = "event.service.nodRmv"

	// EventUpdateService 有一个旧的服务产生了信息的变更（通常是指权重
	EventUpdateService = "event.service.nodUpdate"
)

// IDiscover discover interface
type IDiscover interface {
	Init() error
	Run()
	Close()
}
