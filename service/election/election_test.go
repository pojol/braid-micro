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

	elec, err := New("test", mock.ConsulAddr)
	if err != nil {
		t.Error(err)
	}

	elec.Run()
	time.Sleep(time.Second)
	elec.Close()
}
