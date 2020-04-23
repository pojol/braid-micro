package consul

import (
	"testing"

	"github.com/pojol/braid/mock"
)

func TestHealth(t *testing.T) {

	GetHealthNode(mock.ConsulAddr, "test")

}
