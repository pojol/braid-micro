package consul

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/mock"
)

func TestSession(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(1000)

	sessionName := "test" + strconv.Itoa(r)

	id, err := CreateSession(mock.ConsulAddr, sessionName)
	assert.Equal(t, err, nil)

	err = RefushSession(mock.ConsulAddr, id)
	assert.Equal(t, err, nil)

	ok, err := AcquireLock(mock.ConsulAddr, sessionName, id)
	assert.Equal(t, ok, true)
	assert.Equal(t, err, nil)

	err = ReleaseLock(mock.ConsulAddr, sessionName, id)
	assert.Equal(t, err, nil)

	ok, err = DeleteSession(mock.ConsulAddr, id)
	assert.Equal(t, ok, true)
	assert.Equal(t, err, nil)
}
