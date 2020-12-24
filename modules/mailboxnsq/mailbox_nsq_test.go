package mailboxnsq

import (
	"sync"
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

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	c1, _ := mb.Sub(mailbox.Cluster, "TestClusterShared").Shared()
	defer c1.Exit()
	c2, _ := mb.Sub(mailbox.Cluster, "TestClusterShared").Shared()
	defer c2.Exit()

	go func() {
		for {
			select {
			case <-c1.OnArrived():
				wg.Done()
				c1.Done()
			case <-c2.OnArrived():
				wg.Done()
				c2.Done()
			}
		}
	}()

	mb.Pub(mailbox.Cluster, "TestClusterShared", &mailbox.Message{
		Body: []byte("test msg"),
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		//pass
	case <-time.After(time.Millisecond * 1000):
		t.FailNow()
	}

}

func TestClusterCompetition(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := b.Build("cluster")
	var tick uint64
	var tickmu sync.Mutex

	c1, _ := mb.Sub(mailbox.Cluster, "TestClusterCompetition").Competition()
	defer c1.Exit()
	c2, _ := mb.Sub(mailbox.Cluster, "TestClusterCompetition").Competition()
	defer c2.Exit()

	go func() {
		for {
			select {
			case <-c1.OnArrived():
				tickmu.Lock()
				atomic.AddUint64(&tick, 1)
				c1.Done()
				tickmu.Unlock()
			case <-c2.OnArrived():
				tickmu.Lock()
				atomic.AddUint64(&tick, 1)
				c2.Done()
				tickmu.Unlock()
			}
		}
	}()

	mb.Pub(mailbox.Cluster, "TestClusterCompetition", &mailbox.Message{
		Body: []byte("test msg"),
	})

	time.Sleep(time.Millisecond * 1000)
	tickmu.Lock()
	assert.Equal(t, tick, uint64(1))
	tickmu.Unlock()

}

func TestClusterMailboxParm(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	b.AddOption(WithChannel("parm"))
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))

	mb, err := b.Build("cluster")
	assert.Equal(t, err, nil)
	cm := mb.(*nsqMailbox)

	assert.Equal(t, cm.parm.Channel, "parm")
	assert.Equal(t, cm.parm.LookupAddress, []string{mock.NSQLookupdAddr})
	assert.Equal(t, cm.parm.Address, []string{mock.NsqdAddr})
}
