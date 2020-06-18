package consul

import (
	"testing"

	"github.com/pojol/braid/3rd/log"
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
	GetHealthNode("http://127.0.0.1:8901", "test")

}
