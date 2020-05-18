package consul

import (
	"testing"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestHealth(t *testing.T) {

	l := log.New(log.Config{
		Mode:   log.DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, log.WithSys(log.Config{
		Mode:   log.DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	GetHealthNode(mock.ConsulAddr, "test")
	GetHealthNode("xxxx", "test")

}
