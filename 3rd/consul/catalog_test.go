package consul

import (
	"fmt"
	"testing"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
)

func TestServicesList(t *testing.T) {

	l := log.New(log.Config{
		Mode:   log.DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, log.WithSys(log.Config{
		Mode:   log.DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	_, err := ServicesList(mock.ConsulAddr)
	if err != nil {
		t.Error(err)
	}
	ServicesList("xxx")
}

func TestGetNodeServices(t *testing.T) {

	l := log.New(log.Config{
		Mode:   log.DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, log.WithSys(log.Config{
		Mode:   log.DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	lst, err := GetCatalogServices(mock.ConsulAddr, "redis")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(lst)
	GetCatalogServices("xxx", "redis")
}

func TestCatalog(t *testing.T) {

	l := log.New(log.Config{
		Mode:   log.DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, log.WithSys(log.Config{
		Mode:   log.DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	GetCatalogService(mock.ConsulAddr, "test")
	GetCatalogService("xxx", "test")
}
