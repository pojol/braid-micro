package meta

import (
	"encoding/json"
)

const (
	// 服务发现 - 有服务加入或退出
	TopicDiscoverServiceUpdate = "braid.topic.discover.service_update"
	// 服务发现 - 有新的节点加入服务
	TopicDiscoverServiceNodeAdd = "braid.topic.discover.service_node_add"
	// 服务发现 - 有节点退出服务
	TopicDiscoverServiceNodeRmv = "braid.topic.discover.service_node_rmv"
	// 服务发现 - 有节点信息更新
	TopicDiscoverServiceNodeUpdate = "braid.topic.discover.service_node_update"

	// --------------------------------------------------

	// 链路缓存 - 统计在中连接的总数
	TopicLinkcacheLinkNumber = "braid.topic.linkcache.service_link_number"
	// 链路缓存 - 有用户断开（下线，不再需要持有链路信息
	TopicLinkcacheUnlink = "braid.topic.linkcache.unlink"

	// --------------------------------------------------

	// 选举 - 选举状态变更
	TopicElectionChangeState = "braid.topic.election.change_state"
)

type UpdateMsg struct {
	Nod   Node
	Event string
}

func EncodeUpdateMsg(event string, nod Node) *Message {
	byt, _ := json.Marshal(&UpdateMsg{
		Event: event,
		Nod:   nod,
	})

	return &Message{
		Body: byt,
	}
}

func DecodeUpdateMsg(msg *Message) UpdateMsg {
	dmsg := UpdateMsg{}
	json.Unmarshal(msg.Body, &dmsg)
	return dmsg
}

// LinkNumMsg msg struct
type LinkNumMsg struct {
	ID  string
	Num int
}

// EncodeLinkNumMsg encode linknum msg
func EncodeNumMsg(id string, num int) *Message {
	byt, _ := json.Marshal(&LinkNumMsg{
		ID:  id,
		Num: num,
	})

	return &Message{
		Body: byt,
	}
}

// DecodeLinkNumMsg decode linknum msg
func DecodeNumMsg(msg *Message) LinkNumMsg {
	lnmsg := LinkNumMsg{}
	json.Unmarshal(msg.Body, &lnmsg)
	return lnmsg
}

const (
	// Wait 表示此进程当前处于初始化阶段，还没有具体的选举信息
	EWait int32 = 0 + iota
	// Slave 表示此进程当前处于 从节点 状态，此状态下，elector 会不断进行重试，试图变成新的 主节点（当主节点宕机或退出时
	ESlave
	// Master 表示当前进程正处于 主节点 状态；
	EMaster
)

// StateChangeMsg become master msg
type StateChangeMsg struct {
	State int32
	ID    string
}

// EncodeStateChangeMsg encode
func EncodeStateChangeMsg(state int32, id string) *Message {
	byt, _ := json.Marshal(&StateChangeMsg{
		State: state,
		ID:    id,
	})

	return &Message{
		Body: byt,
	}
}

// DecodeStateChangeMsg decode
func DecodeStateChangeMsg(msg *Message) StateChangeMsg {
	bmmsg := StateChangeMsg{}
	json.Unmarshal(msg.Body, &bmmsg)
	return bmmsg
}
