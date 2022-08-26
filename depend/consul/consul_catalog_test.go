package consul

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const mock_consul_addr = "127.0.0.1:8900"

func TestServicesList(t *testing.T) {

	client := Build(WithAddress([]string{mock_consul_addr}), WithTimeOut(time.Second*10))

	_, err := client.ListServices()
	assert.Equal(t, err, nil)

}

// containsInSlice 判断字符串是否在 slice 中
func containsInSlice(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func TestGetService(t *testing.T) {
	client := Build(WithAddress([]string{mock_consul_addr}), WithTimeOut(time.Second*10))

	services, err := client.ListServices()
	assert.Equal(t, err, nil)

	for _, v := range services {

		/*
			if !containsInSlice(v.Tags, "") {
				continue
			}
		*/
		service, err := client.GetService(v.Name)
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}

		for _, nod := range service.Nodes {
			fmt.Println(nod)
		}

	}

	t.Fail()
}
