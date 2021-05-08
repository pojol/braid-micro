package mailboxnsq

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	mock.Init()

	m.Run()
}

func TestClusterBroadcast(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	b.AddOption(WithNsqLogLv(nsq.LogLevelDebug))
	mb, _ := b.Build("TestClusterBroadcast", log)

	topic := mb.Topic("TestClusterBroadcast")
	channel1 := topic.Sub("Normal_1", mailbox.ScopeCluster)
	channel2 := topic.Sub("Normal_2", mailbox.ScopeCluster)

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	go func() {
		for {
			select {
			case <-channel1.Arrived():
				wg.Done()
			case <-channel2.Arrived():
				wg.Done()
			}
		}
	}()

	go func() {
		wg.Wait()
		close(done)
	}()

	topic.Pub(&mailbox.Message{Body: []byte("test msg")})

	select {
	case <-done:
		// pass
	case <-time.After(time.Second * 2):
		t.FailNow()
	}

}

func TestClusterNotify(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	b.AddOption(WithNsqLogLv(nsq.LogLevelDebug))
	mb, _ := b.Build("TestClusterNotify", log)

	var tick uint64

	topic := mb.Topic("TestClusterNotify")
	channel1 := topic.Sub("Normal", mailbox.ScopeCluster)
	channel2 := topic.Sub("Normal", mailbox.ScopeCluster)

	go func() {
		for {
			select {
			case <-channel1.Arrived():
				atomic.AddUint64(&tick, 1)
			case <-channel2.Arrived():
				atomic.AddUint64(&tick, 1)
			}
		}
	}()

	topic.Pub(&mailbox.Message{Body: []byte("msg")})

	select {
	case <-time.After(time.Second * 2):
		assert.Equal(t, tick, uint64(1))
	}

}

/*
// BenchmarkShared-8   	 2102298	       527 ns/op	      94 B/op	       2 allocs/op
func BenchmarkProcShared(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mbb.Build("BenchmarkProcShared", log)
	topic := "BenchmarkProcShared"
	body := []byte("msg")

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Shared()
	c2, _ := sub.Shared()

	c1.OnArrived(func(msg mailbox.Message) error {
		return nil
	})
	c2.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Pub(mailbox.Proc, topic, &mailbox.Message{Body: body})
	}
}

func BenchmarkProcSharedAsync(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mbb.Build("BenchmarkProcSharedAsync", log)
	topic := "BenchmarkProcSharedAsync"
	body := []byte("msg")

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Shared()
	c2, _ := sub.Shared()

	c1.OnArrived(func(msg mailbox.Message) error {
		return nil
	})
	c2.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mb.PubAsync(mailbox.Proc, topic, &mailbox.Message{Body: body})
		}
	})

}

//BenchmarkCompetition-8   	 3238792	       335 ns/op	      79 B/op	       2 allocs/op
func BenchmarkProcCompetition(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mbb.Build("BenchmarkProcCompetition", log)
	topic := "BenchmarkProcCompetition"
	body := []byte("msg")

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Competition()
	c2, _ := sub.Competition()

	c1.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	c2.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Pub(mailbox.Proc, topic, &mailbox.Message{Body: body})
	}
}

func BenchmarkProcCompetitionAsync(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mbb.Build("BenchmarkProcCompetitionAsync", log)
	topic := "BenchmarkProcCompetitionAsync"
	body := []byte("msg")

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Competition()
	c2, _ := sub.Competition()

	c1.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	c2.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mb.PubAsync(mailbox.Proc, topic, &mailbox.Message{Body: body})
		}
	})
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
*/
