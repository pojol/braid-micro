package consul

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServicesList(t *testing.T) {

	client := Build(WithAddress([]string{"47.103.70.168:8900"}), WithTimeOut(time.Second*10))

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
	client := Build(WithAddress([]string{"47.103.70.168:8900"}), WithTimeOut(time.Second*10))

	services, err := client.ListServices()
	assert.Equal(t, err, nil)

	for _, v := range services {

		if !containsInSlice(v.Tags, "mist_dev") {
			continue
		}

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
