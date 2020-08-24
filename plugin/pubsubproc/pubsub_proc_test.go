package pubsubproc

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/module/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestConsumer(t *testing.T) {

	pb, _ := pubsub.GetBuilder(PubsubName).Build()
	var testTick uint64

	tsub := pb.Sub("test")
	tconsumer := tsub.AddCompetition()
	tconsumer.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("consume topic", "test", string(msg.Body))
		atomic.AddUint64(&testTick, 1)
		return nil
	})

	pb.Pub("test", &pubsub.Message{Body: []byte("test1")})
	pb.Pub("test", &pubsub.Message{Body: []byte("test2")})

	time.Sleep(time.Millisecond * 100)
	tconsumer.Exit()
	pb.Pub("test", pubsub.NewMessage([]byte("test3")))

	time.Sleep(time.Second)
	assert.Equal(t, testTick, uint64(2))
}
