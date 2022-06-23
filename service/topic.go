package service

import (
	"encoding/json"

	"github.com/pojol/braid-go/depend/pubsub"
)

/////////////////////////////// Discover ////////////////////////////////////////

const (
	TopicServiceUpdate = "discover.serviceUpdate"
)

type DiscoverUpdateMsg struct {
	Nod   Node
	Event string
}

func DiscoverEncodeUpdateMsg(event string, nod Node) *pubsub.Message {
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

/////////////////////////////// Elector ////////////////////////////////////////

const (
	TopicElectorChangeState = "elector.changeState"
)

// StateChangeMsg become master msg
type ElectorStateChangeMsg struct {
	State string
}

// EncodeStateChangeMsg encode
func ElectorEncodeStateChangeMsg(state string) *pubsub.Message {
	byt, _ := json.Marshal(&ElectorStateChangeMsg{
		State: state,
	})

	return &pubsub.Message{
		Body: byt,
	}
}

// DecodeStateChangeMsg decode
func ElectorDecodeStateChangeMsg(msg *pubsub.Message) ElectorStateChangeMsg {
	bmmsg := ElectorStateChangeMsg{}
	json.Unmarshal(msg.Body, &bmmsg)
	return bmmsg
}

/////////////////////////////// Linker ///////////////////////////////////////

const (
	// 当前节点连接数事件
	TopicLinkerLinkNum = "linkcache.serviceLinkNum"

	// token 离线事件
	TopicLinkerUnlink = "linkcache.tokenUnlink"
)

// LinkNumMsg msg struct
type LinkerLinkNumMsg struct {
	ID  string
	Num int
}

// EncodeLinkNumMsg encode linknum msg
func LinkerEncodeNumMsg(id string, num int) *pubsub.Message {
	byt, _ := json.Marshal(&LinkerLinkNumMsg{
		ID:  id,
		Num: num,
	})

	return &pubsub.Message{
		Body: byt,
	}
}

// DecodeLinkNumMsg decode linknum msg
func LinkerDecodeNumMsg(msg *pubsub.Message) LinkerLinkNumMsg {
	lnmsg := LinkerLinkNumMsg{}
	json.Unmarshal(msg.Body, &lnmsg)
	return lnmsg
}
