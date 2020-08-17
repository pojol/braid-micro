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

	pb := pubsub.GetBuilder(PubsubName).Build()
	var testTick uint64

	testConsumer := pb.Sub("test")
	testConsumer.AddHandler(func(msg *pubsub.Message) error {
		fmt.Println("consume topic", "test", msg.Body)
		atomic.AddUint64(&testTick, 1)
		return nil
	})

	pb.Pub("test", pubsub.NewMessage("test1"))
	pb.Pub("test", pubsub.NewMessage("test2"))

	time.Sleep(time.Millisecond * 100)
	testConsumer.Exit()
	pb.Pub("test", pubsub.NewMessage("test3"))

	time.Sleep(time.Second)
	assert.Equal(t, testTick, uint64(2))
}
