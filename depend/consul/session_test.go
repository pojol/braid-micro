package consul

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSession(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(1000)

	sessionName := "test" + strconv.Itoa(r)

	client := BuildWithOption(WithAddress([]string{mock_consul_addr}), WithTimeOut(time.Second*10))

	id, err := client.CreateSession(sessionName)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)

	err = client.RefreshSession(id)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)

	err = client.DeleteSession(id)
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)
}
