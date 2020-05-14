package consul

import (
	"testing"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestHealth(t *testing.T) {

	l := log.New("test")
	l.Init()

	GetHealthNode(mock.ConsulAddr, "test")
	GetHealthNode("xxxx", "test")

}
