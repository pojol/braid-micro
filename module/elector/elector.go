package elector

import (
	"encoding/json"

	"github.com/pojol/braid-go/module/pubsub"
)

const (
	TopicChangeState = "elector.local.changeState"
)

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
}

// EncodeStateChangeMsg encode
func EncodeStateChangeMsg(state int32) *pubsub.Message {
	byt, _ := json.Marshal(&StateChangeMsg{
		State: state,
	})

	return &pubsub.Message{
		Body: byt,
	}
}

// DecodeStateChangeMsg decode
func DecodeStateChangeMsg(msg *pubsub.Message) StateChangeMsg {
	bmmsg := StateChangeMsg{}
	json.Unmarshal(msg.Body, &bmmsg)
	return bmmsg
}

// IElector election interface
type IElector interface {
	Init() error
	Run()
	Close()
}
