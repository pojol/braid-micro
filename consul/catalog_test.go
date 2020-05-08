package consul

import (
	"fmt"
	"testing"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestServicesList(t *testing.T) {

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	_, err := ServicesList(mock.ConsulAddr)
	if err != nil {
		t.Error(err)
	}
	ServicesList("xxx")
}

func TestGetNodeServices(t *testing.T) {

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	lst, err := GetCatalogServices(mock.ConsulAddr, "redis")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(lst)
	GetCatalogServices("xxx", "redis")
}

func TestCatalog(t *testing.T) {

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	GetCatalogService(mock.ConsulAddr, "test")
	GetCatalogService("xxx", "test")
}
