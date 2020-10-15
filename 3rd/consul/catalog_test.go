package consul

import (
	"testing"

	"github.com/pojol/braid/mock"
)

func TestServicesList(t *testing.T) {

	_, err := ServicesList(mock.ConsulAddr)
	if err != nil {
		t.Error(err)
	}
	ServicesList("xxx")
}

func TestGetNodeServices(t *testing.T) {

	_, err := GetCatalogServices(mock.ConsulAddr, "redis")
	if err != nil {
		t.Error(err)
	}

	GetCatalogServices("xxx", "redis")
}

func TestCatalog(t *testing.T) {

	GetCatalogService(mock.ConsulAddr, "test")
	GetCatalogService("xxx", "test")
}
