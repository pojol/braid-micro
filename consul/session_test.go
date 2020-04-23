package consul

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/mock"
)

func TestSession(t *testing.T) {

	id, err := CreateSession(mock.ConsulAddr, "test")
	assert.Equal(t, err, nil)

	err = RefushSession(mock.ConsulAddr, id)
	assert.Equal(t, err, nil)

	ok, err := AcquireLock(mock.ConsulAddr, "test", id)
	assert.Equal(t, ok, true)
	assert.Equal(t, err, nil)

	err = ReleaseLock(mock.ConsulAddr, "test", id)
	assert.Equal(t, err, nil)

	ok, err = DeleteSession(mock.ConsulAddr, id)
	assert.Equal(t, ok, true)
	assert.Equal(t, err, nil)
}
