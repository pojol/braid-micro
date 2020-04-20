package election

import (
	"testing"
	"time"

	"github.com/pojol/braid/mock"
)

func TestElection(t *testing.T) {

	mock.Init()

	e := New()
	e.Init(Config{
		Address: mock.ConsulAddr,
		Name:    "test",
	})

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}
