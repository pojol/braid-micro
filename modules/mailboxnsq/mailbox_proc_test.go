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
	var tickmu sync.Mutex

	mb.RegistTopic("TestProcNotify", mailbox.ScopeProc)
	topic := mb.GetTopic("TestProcNotify")
	channel1 := topic.Sub("Normal")
	channel2 := topic.Sub("Normal")

	go func() {
		for {
			select {
			case <-channel1.Arrived():
				tickmu.Lock()
				atomic.AddUint64(&tick, 1)
				tickmu.Unlock()
			case <-channel2.Arrived():
				tickmu.Lock()
				atomic.AddUint64(&tick, 1)
				tickmu.Unlock()
			}
		}
	}()

	topic.Pub(&mailbox.Message{Body: []byte("msg")})

	select {
	case <-time.After(time.Second):
		tickmu.Lock()
		assert.Equal(t, tick, uint64(1))
		tickmu.Unlock()
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

	mb.RegistTopic("BenchmarkTestProc", mailbox.ScopeProc)
	topic := mb.GetTopic("BenchmarkTestProc")
	c1 := topic.Sub("Normal")
	c2 := topic.Sub("Normal")

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
