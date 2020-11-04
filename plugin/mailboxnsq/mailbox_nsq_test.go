package mailboxnsq

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/mailbox"
	"github.com/stretchr/testify/assert"
)

func TestClusterShared(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := b.Build("cluster")
	var tick uint64

	c1, _ := mb.ClusterSub("TestClusterShared").AddShared()
	c1.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&tick, 1)
		return nil
	})

	c2, _ := mb.ClusterSub("TestClusterShared").AddShared()
	c2.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&tick, 1)
		return nil
	})

	mb.ClusterPub("TestClusterShared", &mailbox.Message{
		Body: []byte("test msg"),
	})
	time.Sleep(time.Millisecond * 500)

	assert.Equal(t, tick, 2)
}

func TestClusterCompetition(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := b.Build("cluster")
	var tick uint64

	c1, _ := mb.ClusterSub("TestClusterCompetition").AddCompetition()
	c1.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&tick, 1)
		return nil
	})

	c2, _ := mb.ClusterSub("TestClusterCompetition").AddCompetition()
	c2.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&tick, 1)
		return nil
	})

	mb.ClusterPub("TestClusterCompetition", &mailbox.Message{
		Body: []byte("test msg"),
	})
	time.Sleep(time.Millisecond * 500)

	assert.Equal(t, tick, 1)
}

func TestClusterMailboxParm(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	b.AddOption(WithChannel("parm"))
	b.AddOption(WithLookupAddr([]string{"127.0.0.1"}))
	b.AddOption(WithNsqdAddr([]string{"127.0.0.2"}))

	mb, _ := b.Build("cluster")
	cm := mb.(*nsqMailbox)

	assert.Equal(t, cm.parm.Channel, "parm")
	assert.Equal(t, cm.parm.LookupAddress, []string{"127.0.0.1"})
	assert.Equal(t, cm.parm.Address, []string{"127.0.0.2"})
}
