package consul

import (
	"fmt"
	"testing"
	"time"

	"github.com/pojol/braid-go/internal/utils"
	"github.com/stretchr/testify/assert"
)

const mock_consul_addr = "47.103.70.168:8900"

func TestServicesList(t *testing.T) {

	client := Build(WithAddress([]string{mock_consul_addr}), WithTimeOut(time.Second*10))

	_, err := client.CatalogListServices()
	assert.Equal(t, err, nil)

}

func TestGetService(t *testing.T) {
	client := Build(WithAddress([]string{mock_consul_addr}), WithTimeOut(time.Second*10))

	services, err := client.CatalogListServices()
	assert.Equal(t, err, nil)

	for _, v := range services {

		if !utils.ContainsInSlice(v.Tags, "mist_dev") {
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

	t.Fail()
}
