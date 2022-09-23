package discoverconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/discover"
)

func TestMain(m *testing.M) {
	mock.Init()

	m.Run()
}

func TestDiscover(t *testing.T) {

	d := BuildWithOption(
		"base_dev",
		blog.BuildWithOption(),
		nil,
		consul.BuildWithOption(consul.WithAddress([]string{mock.ConsulAddr})),
		discover.WithSyncServiceInterval(time.Millisecond*100),
		discover.WithSyncServiceWeightInterval(time.Millisecond*100),
		discover.WithBlacklist([]string{"gate"}),
		discover.WithTag("braid"),
	)

	d.Init()
	d.Run()

	time.Sleep(time.Second)
	d.Close()

	t.FailNow()
}
