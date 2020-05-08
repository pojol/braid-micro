package consul

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestSession(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(1000)

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	sessionName := "test" + strconv.Itoa(r)

	id, err := CreateSession(mock.ConsulAddr, sessionName)
	assert.Equal(t, err, nil)
	CreateSession("xxx", sessionName)

	err = RefushSession(mock.ConsulAddr, id)
	assert.Equal(t, err, nil)
	RefushSession("xxx", id)

	ok, err := AcquireLock(mock.ConsulAddr, sessionName, id)
	assert.Equal(t, ok, true)
	assert.Equal(t, err, nil)
	AcquireLock("xxx", sessionName, id)

	err = ReleaseLock(mock.ConsulAddr, sessionName, id)
	assert.Equal(t, err, nil)
	ReleaseLock("xxx", sessionName, id)

	ok, err = DeleteSession(mock.ConsulAddr, id)
	assert.Equal(t, ok, true)
	assert.Equal(t, err, nil)
	DeleteSession("xxx", id)
}
