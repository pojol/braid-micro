package consul

import (
	"testing"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestHealth(t *testing.T) {

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	GetHealthNode(mock.ConsulAddr, "test")
	GetHealthNode("xxxx", "test")

}
