package discoverconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/internal/pubsubnsq"
	"github.com/pojol/braid-go/module/pubsub"
)

func TestMain(m *testing.M) {
	mock.Init()

	m.Run()
}

func TestDiscover(t *testing.T) {

	log := blog.BuildWithOption()

	d := BuildWithOption(
		"base_dev",
		log,
		pubsubnsq.BuildWithOption("", log, pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr})),
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
