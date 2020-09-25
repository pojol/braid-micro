package pubsubproc

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/module/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestConsumer(t *testing.T) {

	pb, _ := pubsub.GetBuilder(PubsubName).Build("TestConsumer")
	var testTick uint64

	tsub := pb.Sub("TestConsumer")
	tconsumer := tsub.AddCompetition()
	tconsumer.OnArrived(func(msg *pubsub.Message) error {
		atomic.AddUint64(&testTick, 1)
		return nil
	})

	pb.Pub("TestConsumer", &pubsub.Message{Body: []byte("test1")})
	pb.Pub("TestConsumer", &pubsub.Message{Body: []byte("test2")})

	time.Sleep(time.Millisecond * 100)
	tconsumer.Exit()
	pb.Pub("TestConsumer", pubsub.NewMessage([]byte("test3")))

	time.Sleep(time.Second)
	assert.Equal(t, testTick, uint64(2))
}
