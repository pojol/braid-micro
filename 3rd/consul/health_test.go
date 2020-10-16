package consul

import (
	"testing"

	"github.com/pojol/braid/mock"
)

func TestHealth(t *testing.T) {

	GetHealthNode(mock.ConsulAddr, "test")
	GetHealthNode("http://127.0.0.1:8901", "test")

}
