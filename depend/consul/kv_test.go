package consul

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/pojol/braid-go/mock"
	"github.com/stretchr/testify/assert"
)

func TestKv(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(1000)

	sessionName := "test" + strconv.Itoa(r)

	client := BuildWithOption(WithAddress([]string{mock.ConsulAddr}), WithTimeOut(time.Second*10))

	id, err := client.CreateSession(sessionName)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)

	client.AcquireLock(sessionName, id)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)

	client.DeleteSession(id)
}
