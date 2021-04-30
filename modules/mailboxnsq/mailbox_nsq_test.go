package mailboxnsq

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestClusterShared(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := b.Build("cluster", log)

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	c1, _ := mb.Sub(mailbox.Cluster, "TestClusterShared").Shared()
	defer c1.Exit()
	c2, _ := mb.Sub(mailbox.Cluster, "TestClusterShared").Shared()
	defer c2.Exit()

	c1.OnArrived(func(msg mailbox.Message) error {
		wg.Done()
		return nil
	})

	c2.OnArrived(func(msg mailbox.Message) error {
		wg.Done()
		return nil
	})

	mb.PubAsync(mailbox.Cluster, "TestClusterShared", &mailbox.Message{
		Body: []byte("test msg"),
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		//pass
	case <-time.After(time.Second * 10):
		fmt.Println("TestClusterShared test time out")
		t.FailNow()
	}

}

func TestClusterCompetition(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := b.Build("cluster", log)
	var tick uint64
	var tickmu sync.Mutex

	c1, _ := mb.Sub(mailbox.Cluster, "TestClusterCompetition").Competition()
	defer c1.Exit()
	c2, _ := mb.Sub(mailbox.Cluster, "TestClusterCompetition").Competition()
	defer c2.Exit()

	c1.OnArrived(func(msg mailbox.Message) error {
		tickmu.Lock()
		atomic.AddUint64(&tick, 1)
		tickmu.Unlock()
		return nil
	})

	c2.OnArrived(func(msg mailbox.Message) error {
		tickmu.Lock()
		atomic.AddUint64(&tick, 1)
		tickmu.Unlock()
		return nil
	})

	mb.PubAsync(mailbox.Cluster, "TestClusterCompetition", &mailbox.Message{
		Body: []byte("test msg"),
	})

	time.Sleep(time.Second * 10)
	tickmu.Lock()
	assert.Equal(t, tick, uint64(1))
	tickmu.Unlock()

}

func TestClusterMailboxParm(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	b.AddOption(WithChannel("parm"))
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))

	mb, err := b.Build("cluster", log)
	assert.Equal(t, err, nil)
	cm := mb.(*nsqMailbox)

	assert.Equal(t, cm.parm.Channel, "parm")
	assert.Equal(t, cm.parm.LookupAddress, []string{mock.NSQLookupdAddr})
	assert.Equal(t, cm.parm.Address, []string{mock.NsqdAddr})
}
