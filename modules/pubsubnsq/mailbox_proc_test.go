package pubsubnsq

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/moduleparm"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestProcNotify(t *testing.T) {

	log := module.GetBuilder(zaplogger.Name).Build("TestProcNotify").(logger.ILogger)
	mb := module.GetBuilder(Name).Build("TestProcNotify", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	var tick uint64

	mb.RegistTopic("TestProcNotify", pubsub.ScopeProc)
	topic := mb.GetTopic("TestProcNotify")
	channel1 := topic.Sub("Normal")
	channel2 := topic.Sub("Normal")

	channel1.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})
	channel2.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})

	topic.Pub(&pubsub.Message{Body: []byte("msg")})

	for {
		<-time.After(time.Second)
		assert.Equal(t, atomic.LoadUint64(&tick), uint64(1))
		break
	}
}

func TestProcExit(t *testing.T) {

	log := module.GetBuilder(zaplogger.Name).Build("TestProcExit").(logger.ILogger)
	mb := module.GetBuilder(Name).Build("TestProcExit", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	var tick uint64

	mb.RegistTopic("TestProcExit", pubsub.ScopeProc)
	topic := mb.GetTopic("TestProcExit")
	channel1 := topic.Sub("Normal_1")
	channel2 := topic.Sub("Normal_2")

	channel1.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})
	channel2.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})

	topic.RemoveChannel("Normal_1")

	topic.Pub(&pubsub.Message{Body: []byte("msg")})

	for {
		<-time.After(time.Second * 2)
		assert.Equal(t, atomic.LoadUint64(&tick), uint64(1))

		err := mb.RemoveTopic("TestProcExit")
		assert.Equal(t, err, nil)

		err = topic.Pub(&pubsub.Message{Body: []byte("msg")})
		assert.NotEqual(t, err, nil)
		break
	}
}

func TestProcBroadcast(t *testing.T) {

	log := module.GetBuilder(zaplogger.Name).Build("TestProcBroadcast").(logger.ILogger)
	mb := module.GetBuilder(Name).Build("TestProcBroadcast", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	mb.RegistTopic("TestProcBroadcast", pubsub.ScopeProc)
	topic := mb.GetTopic("TestProcBroadcast")
	channel1 := topic.Sub("Broadcast_Consumer1")
	channel2 := topic.Sub("Broadcast_Consumer2")

	channel1.Arrived(func(msg *pubsub.Message) {
		wg.Done()
	})
	channel2.Arrived(func(msg *pubsub.Message) {
		wg.Done()
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	topic.Pub(&pubsub.Message{Body: []byte("msg")})

	select {
	case <-done:
		// pass
	case <-time.After(time.Second):
		t.FailNow()
	}
}

// 9257995	       130 ns/op	      77 B/op	       2 allocs/op
func BenchmarkTestProc(b *testing.B) {

	log := module.GetBuilder(zaplogger.Name).Build("BenchmarkTestProc", zaplogger.WithLv(logger.ERROR)).(logger.ILogger)
	mb := module.GetBuilder(Name).Build("BenchmarkTestProc", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	body := []byte("msg")

	mb.RegistTopic("BenchmarkTestProc", pubsub.ScopeProc)
	topic := mb.GetTopic("BenchmarkTestProc")
	c1 := topic.Sub("Normal")
	c2 := topic.Sub("Normal")

	c1.Arrived(func(msg *pubsub.Message) {

	})
	c2.Arrived(func(msg *pubsub.Message) {

	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topic.Pub(&pubsub.Message{Body: body})
		}
	})
}
