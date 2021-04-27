package elector

import (
	"encoding/json"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/mailbox"
)

const (
	// ElectorStateChange topic_elector_state
	ElectorStateChange = "topic_elector_state"
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
	EWait   = "elector_wait"
	ESlave  = "elector_slave"
	EMaster = "elector_master"
)

// IElection election interface
type IElection interface {
	module.IModule
}
