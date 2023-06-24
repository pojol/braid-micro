package pubsubnsq

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/meta"
	"github.com/stretchr/testify/assert"
)

func TestProcNotify(t *testing.T) {

	psc := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		blog.BuildWithDefaultOption(),
		WithLookupAddr([]string{}),
		WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)

	var tick uint64

	topic := psc.GetTopic("TestProcNotify")
	channel1, _ := topic.Sub(context.TODO(), "Normal")
	channel2, _ := topic.Sub(context.TODO(), "Normal")

	defer topic.Close()

	channel1.Arrived(func(msg *meta.Message) error {
		atomic.AddUint64(&tick, 1)
		return nil
	})
	channel2.Arrived(func(msg *meta.Message) error {
		atomic.AddUint64(&tick, 1)
		return nil
	})

	topic.Pub(context.TODO(), &meta.Message{Body: []byte("msg")})

	for {
		<-time.After(time.Second)
		assert.Equal(t, atomic.LoadUint64(&tick), uint64(1))
		break
	}
}

func TestProcExit(t *testing.T) {

	ps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		blog.BuildWithDefaultOption(),
		WithLookupAddr([]string{}),
		WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)

	var tick uint64

	topic := ps.GetTopic("TestProcExit")
	channel1, _ := topic.Sub(context.TODO(), "Normal_1")
	channel2, _ := topic.Sub(context.TODO(), "Normal_2")
	defer topic.Close()

	channel1.Arrived(func(msg *meta.Message) error {
		atomic.AddUint64(&tick, 1)
		return nil
	})
	channel2.Arrived(func(msg *meta.Message) error {
		atomic.AddUint64(&tick, 1)
		return nil
	})

	err := channel1.Close()
	assert.Equal(t, err, nil)

	topic.Pub(context.TODO(), &meta.Message{Body: []byte("msg")})

	for {
		<-time.After(time.Second * 2)
		assert.Equal(t, atomic.LoadUint64(&tick), uint64(2))

		err := topic.Close()
		assert.Equal(t, err, nil)

		err = topic.Pub(context.TODO(), &meta.Message{Body: []byte("msg")})
		assert.NotEqual(t, err, nil)
		break
	}
}

func TestProcBroadcast(t *testing.T) {

	ps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		blog.BuildWithDefaultOption(),
		WithLookupAddr([]string{}),
		WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	topic := ps.GetTopic("TestProcBroadcast")
	channel1, _ := topic.Sub(context.TODO(), "Broadcast_Consumer1")
	channel2, _ := topic.Sub(context.TODO(), "Broadcast_Consumer2")
	defer topic.Close()

	channel1.Arrived(func(msg *meta.Message) error {
		wg.Done()
		return nil
	})
	channel2.Arrived(func(msg *meta.Message) error {
		wg.Done()
		return nil
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	topic.Pub(context.TODO(), &meta.Message{Body: []byte("msg")})

	select {
	case <-done:
		// pass
	case <-time.After(time.Second):
		t.FailNow()
	}
}

// 9257995	       130 ns/op	      77 B/op	       2 allocs/op
func BenchmarkTestProc(b *testing.B) {

	ps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		blog.BuildWithDefaultOption(),
		WithLookupAddr([]string{}),
		WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)

	body := []byte("msg")

	topic := ps.GetTopic("BenchmarkTestProc")
	c1, _ := topic.Sub(context.TODO(), "Normal")
	c2, _ := topic.Sub(context.TODO(), "Normal")

	c1.Arrived(func(msg *meta.Message) error {
		return nil
	})
	c2.Arrived(func(msg *meta.Message) error {
		return nil
	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topic.Pub(context.TODO(), &meta.Message{Body: body})
		}
	})
}
