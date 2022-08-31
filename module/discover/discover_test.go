package discover

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/mock"
)

func TestMain(m *testing.M) {
	mock.Init()

	m.Run()
}

func TestDiscover(t *testing.T) {

	d := Build("mist_dev", nil, consul.Build(consul.WithAddress([]string{"47.103.70.168:8900"})),
		WithSyncServiceInterval(time.Millisecond*100),
		WithSyncServiceWeightInterval(time.Millisecond*100),
		WithBlacklist([]string{"gate"}),
	)

	d.Init()
	d.Run()

	time.Sleep(time.Second)
	d.Close()

	t.FailNow()
}

/*
func TestParm(t *testing.T) {
	b := module.GetBuilder(Name)

	blog.New(blog.NewWithDefault())
	mb := module.GetBuilder(pubsub.Name).Build("TestParm").(pubsub.IPubsub)

	b.AddModuleOption(WithConsulAddr(mock.ConsulAddr))
	b.AddModuleOption(WithTag("TestParm"))
	b.AddModuleOption(WithBlacklist([]string{"gate"}))
	b.AddModuleOption(WithSyncServiceInterval(time.Second))
	b.AddModuleOption(WithSyncServiceWeightInterval(time.Second))

	discv := b.Build("test",
		moduleparm.WithPubsub(mb)).(*consulDiscover)

	assert.Equal(t, discv.parm.Address, mock.ConsulAddr)
	assert.Equal(t, discv.parm.Tag, "TestParm")
	assert.Equal(t, discv.parm.Blacklist, []string{"gate"})
	assert.Equal(t, discv.parm.SyncServicesInterval, time.Second)
	assert.Equal(t, discv.parm.SyncServiceWeightInterval, time.Second)
}
*/
