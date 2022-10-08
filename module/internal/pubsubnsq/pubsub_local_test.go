package pubsubnsq

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestProcNotify(t *testing.T) {

	mb := BuildWithOption(
		"TestProcNotify",
		blog.BuildWithOption(),
	)

	var tick uint64

	topic := mb.LocalTopic("TestProcNotify")
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

	mb := BuildWithOption(
		"TestProcExit",
		blog.BuildWithOption(),
	)

	var tick uint64

	topic := mb.LocalTopic("TestProcExit")
	channel1 := topic.Sub("Normal_1")
	channel2 := topic.Sub("Normal_2")

	channel1.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})
	channel2.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})

	err := channel1.Close()
	assert.Equal(t, err, nil)

	topic.Pub(&pubsub.Message{Body: []byte("msg")})

	for {
		<-time.After(time.Second * 2)
		assert.Equal(t, atomic.LoadUint64(&tick), uint64(1))

		err := topic.Close()
		assert.Equal(t, err, nil)

		err = topic.Pub(&pubsub.Message{Body: []byte("msg")})
		assert.NotEqual(t, err, nil)
		break
	}
}

func TestProcBroadcast(t *testing.T) {

	mb := BuildWithOption(
		"TestProcBroadcast",
		blog.BuildWithOption(),
	)

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	topic := mb.LocalTopic("TestProcBroadcast")
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

	mb := BuildWithOption(
		"BenchmarkTestProc",
		blog.BuildWithOption(),
	)

	body := []byte("msg")

	topic := mb.LocalTopic("BenchmarkTestProc")
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
