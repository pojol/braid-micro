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

	topic := mb.Topic("TestProcNotify")
	channel1 := topic.Channel("Normal", mailbox.ScopeProc)
	channel2 := topic.Channel("Normal", mailbox.ScopeProc)

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
	case <-time.After(time.Second):
		assert.Equal(t, tick, uint64(1))
	}
}

func TestProcBoradcast(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := b.Build("TestProcBoradcast", log)

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	topic := mb.Topic("TestProcBoradcast")
	channel1 := topic.Channel("Boradcast_Consumer1", mailbox.ScopeProc)
	channel2 := topic.Channel("Boradcast_Consumer2", mailbox.ScopeProc)

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

	topic := mb.Topic("BenchmarkTestProc")
	c1 := topic.Channel("Normal", mailbox.ScopeProc)
	c2 := topic.Channel("Normal", mailbox.ScopeProc)

	go func() {
		for {
			select {
			case <-c1.Arrived():
			case <-c2.Arrived():
			}
		}
	}()

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topic.Pub(&mailbox.Message{Body: body})
		}
	})
}
