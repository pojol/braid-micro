package consul

import (
	"testing"

	"github.com/pojol/braid-go/mock"
	"github.com/stretchr/testify/assert"
)

func TestGetConsulLeader(t *testing.T) {

	mock.Init()

	leader, err := GetConsulLeader(mock.ConsulAddr)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, leader, "")
}
