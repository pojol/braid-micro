package elector

import (
	"testing"

	"github.com/pojol/braid-go/mock"
)

func TestMain(m *testing.M) {
	mock.Init()
	m.Run()
}

/*
func TestElection(t *testing.T) {

	blog.New(blog.NewWithDefault())
	mb := module.GetBuilder(pubsubnsq.Name).Build("TestElection").(pubsub.IPubsub)

	eb := module.GetBuilder(Name)
	eb.AddModuleOption(WithConsulAddr(mock.ConsulAddr))

	e := eb.Build("test_elector_with_consul",
		moduleparm.WithPubsub(mb)).(elector.IElector)

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}
*/

/*
func TestParm(t *testing.T) {

	blog.New(blog.NewWithDefault())
	mb := module.GetBuilder(pubsubnsq.Name).Build("TestParm").(pubsub.IPubsub)

	eb := module.GetBuilder(Name)
	eb.AddModuleOption(WithConsulAddr(mock.ConsulAddr))
	eb.AddModuleOption(WithLockTick(time.Second))
	eb.AddModuleOption(WithSessionTick(time.Second))

	e := eb.Build("test_elector_with_consul",
		moduleparm.WithPubsub(mb)).(*consulElection)

	assert.Equal(t, e.parm.ConsulAddr, mock.ConsulAddr)
	assert.Equal(t, e.parm.LockTick, time.Second)
	assert.Equal(t, e.parm.RefushSessionTick, time.Second)
}
*/
