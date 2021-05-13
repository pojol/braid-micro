package mailboxnsq

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestProcNotify(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := b.Build("TestProcNotify", log)

	var tick uint64

	mb.RegistTopic("TestProcNotify", mailbox.ScopeProc)
	topic := mb.GetTopic("TestProcNotify")
	channel1 := topic.Sub("Normal")
	channel2 := topic.Sub("Normal")

	channel1.Arrived(func(msg *mailbox.Message) {
		atomic.AddUint64(&tick, 1)
	})
	channel2.Arrived(func(msg *mailbox.Message) {
		atomic.AddUint64(&tick, 1)
	})

	topic.Pub(&mailbox.Message{Body: []byte("msg")})

	select {
	case <-time.After(time.Second):
		assert.Equal(t, atomic.LoadUint64(&tick), uint64(1))
	}
}

func TestProcExit(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := b.Build("TestProcExit", log)

	var tick uint64

	mb.RegistTopic("TestProcExit", mailbox.ScopeProc)
	topic := mb.GetTopic("TestProcExit")
	channel1 := topic.Sub("Normal_1")
	channel2 := topic.Sub("Normal_2")

	channel1.Arrived(func(msg *mailbox.Message) {
		atomic.AddUint64(&tick, 1)
	})
	channel2.Arrived(func(msg *mailbox.Message) {
		atomic.AddUint64(&tick, 1)
	})

	topic.RemoveChannel("Normal_1")

	topic.Pub(&mailbox.Message{Body: []byte("msg")})

	err := topic.Pub(&mailbox.Message{Body: []byte("msg")})
	assert.NotEqual(t, err, nil)

	select {
	case <-time.After(time.Second * 2):
		err = mb.RemoveTopic("TestProcExit")
		assert.Equal(t, err, nil)

		assert.Equal(t, atomic.LoadUint64(&tick), uint64(1))
	}
}

func TestProcBoradcast(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := b.Build("TestProcBoradcast", log)

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	mb.RegistTopic("TestProcBoradcast", mailbox.ScopeProc)
	topic := mb.GetTopic("TestProcBoradcast")
	channel1 := topic.Sub("Boradcast_Consumer1")
	channel2 := topic.Sub("Boradcast_Consumer2")

	channel1.Arrived(func(msg *mailbox.Message) {
		wg.Done()
	})
	channel2.Arrived(func(msg *mailbox.Message) {
		wg.Done()
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	topic.Pub(&mailbox.Message{Body: []byte("msg")})

	select {
	case <-done:
		// pass
	case <-time.After(time.Second):
		t.FailNow()
	}
}

// 9257995	       130 ns/op	      77 B/op	       2 allocs/op
func BenchmarkTestProc(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mbb.Build("BenchmarkTestProc", log)
	body := []byte("msg")

	mb.RegistTopic("BenchmarkTestProc", mailbox.ScopeProc)
	topic := mb.GetTopic("BenchmarkTestProc")
	c1 := topic.Sub("Normal")
	c2 := topic.Sub("Normal")

	c1.Arrived(func(msg *mailbox.Message) {

	})
	c2.Arrived(func(msg *mailbox.Message) {

	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topic.Pub(&mailbox.Message{Body: body})
		}
	})
}
