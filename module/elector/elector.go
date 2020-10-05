package elector

import (
	"github.com/pojol/braid/module"
)

const (
	// BecomeMaster topic_become_master
	BecomeMaster = "topic_become_master"
)

// IElection election interface
type IElection interface {
	module.IModule
}
