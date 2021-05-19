// 接口文件 elector 选举，主要用于分布式系统中的选举
//
// 这个模块会创建 ChangeState Topic，通过这个 Topic 发布当前进程所处于的状态
package elector

import (
	"encoding/json"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/mailbox"
)

const (
	ChangeState = "elector.changeState"
)

// StateChangeMsg become master msg
type StateChangeMsg struct {
	State string
}

// EncodeStateChangeMsg encode
func EncodeStateChangeMsg(state string) *mailbox.Message {
	byt, _ := json.Marshal(&StateChangeMsg{
		State: state,
	})

	return &mailbox.Message{
		Body: byt,
	}
}

// DecodeStateChangeMsg decode
func DecodeStateChangeMsg(msg *mailbox.Message) StateChangeMsg {
	bmmsg := StateChangeMsg{}
	json.Unmarshal(msg.Body, &bmmsg)
	return bmmsg
}

// state
const (
	// Wait 表示此进程当前处于初始化阶段，还没有具体的选举信息
	EWait = "elector_wait"

	// Slave 表示此进程当前处于 从节点 状态，此状态下，elector 会不断进行重试，试图变成新的 主节点（当主节点宕机或退出时
	ESlave = "elector_slave"

	// Master 表示当前进程正处于 主节点 状态；
	EMaster = "elector_master"
)

// IElection election interface
type IElection interface {
	module.IModule
}
