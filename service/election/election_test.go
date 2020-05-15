package election

import (
	"testing"
	"time"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestElection(t *testing.T) {

	mock.Init()

	l := log.New("test")
	l.Init()

	e := New(mock.ConsulAddr, "test")
	e.Init()

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}
