package consul

import (
	"fmt"
	"testing"
	"time"

	"github.com/pojol/braid-go/internal/utils"
	"github.com/pojol/braid-go/mock"
	"github.com/stretchr/testify/assert"
)

func TestServicesList(t *testing.T) {

	client := BuildWithOption(WithAddress([]string{mock.ConsulAddr}), WithTimeOut(time.Second*10))

	_, err := client.CatalogListServices()
	assert.Equal(t, err, nil)

}

func TestGetService(t *testing.T) {
	client := BuildWithOption(WithAddress([]string{mock.ConsulAddr}), WithTimeOut(time.Second*10))

	services, err := client.CatalogListServices()
	assert.Equal(t, err, nil)

	for _, v := range services {

		if !utils.ContainsInSlice(v.Tags, "braid") {
			continue
		}

		fmt.Println("service =>", v.Name, "tags =>", v.Tags)
		service, err := client.CatalogGetService(v.Name)
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}

		for _, nod := range service.Nodes {
			fmt.Println(nod.ID, nod.Address, nod.Port, nod.Metadata)
		}

	}

}
