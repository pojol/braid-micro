package consul

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid-go/mock"
)

func TestSession(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(1000)

	sessionName := "test" + strconv.Itoa(r)

	id, err := CreateSession(mock.ConsulAddr, sessionName)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)
	CreateSession("xxx", sessionName)

	err = RefushSession(mock.ConsulAddr, id)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)
	RefushSession("xxx", id)

	_, err = AcquireLock(mock.ConsulAddr, sessionName, id)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)
	AcquireLock("xxx", sessionName, id)

	err = ReleaseLock(mock.ConsulAddr, sessionName, id)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)
	ReleaseLock("xxx", sessionName, id)

	_, err = DeleteSession(mock.ConsulAddr, id)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)
	DeleteSession("xxx", id)
}
