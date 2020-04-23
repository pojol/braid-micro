package discover

import (
	"testing"

	"github.com/pojol/braid/mock"
)

func TestDiscover(t *testing.T) {

	mock.Init()

	d := New()
	d.Init(Config{
		Interval:      1000,
		ConsulAddress: mock.ConsulAddr,
	})

	d.Run()
	d.Close()
}
