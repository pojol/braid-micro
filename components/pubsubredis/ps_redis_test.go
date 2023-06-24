package pubsubredis

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/depends/bredis"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	mock.Init()

	code := m.Run()
	// 清理测试环境

	os.Exit(code)
}

func TestTopic(t *testing.T) {

	log := blog.BuildWithDefaultOption()
	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	ctx := context.TODO()
	var err error

	topic := redisps.GetTopic("test.topic.1")
	err = topic.Pub(ctx, nil)
	assert.NotEqual(t, err, nil) // redis: nil command
	defer topic.Close()

	err = topic.Pub(ctx, &meta.Message{})
	assert.Equal(t, err, nil)

	channel, err := topic.Sub(ctx, "channel.1")
	assert.Equal(t, err, nil) // redis: nil command

	channel.Arrived(func(msg *meta.Message) error {
		assert.Equal(t, msg.Body, []byte("test msg"))
		return nil
	})
	defer channel.Close()

	topic.Pub(ctx, &meta.Message{Body: []byte("test msg")})

}

// 模拟单个消费者
func TestSingleConsumer(t *testing.T) {

	msglst := []string{"a", "b", "c", "d", "e", "f", "g"}
	consumerMsgLst := []string{"a", "b", "c", "d", "e", "f", "g"}

	log := blog.BuildWithDefaultOption()
	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	ctx := context.TODO()
	var err error

	topic := redisps.GetTopic("test.topic.sigle")

	channel, err := topic.Sub(ctx, "channel.1")
	assert.Equal(t, err, nil) // redis: nil command

	channel.Arrived(func(msg *meta.Message) error {
		assert.Equal(t, msg.Body, []byte(consumerMsgLst[0]))
		consumerMsgLst = consumerMsgLst[1:]
		return nil
	})

	for _, v := range msglst {
		topic.Pub(ctx, &meta.Message{Body: []byte(v)})
	}

	time.Sleep(time.Second)
	channel.Close()
	topic.Close()

	cnt, err := rediscli.Exists(context.TODO(), "test.topic.sigle").Result()
	assert.Equal(t, err, nil)
	assert.Equal(t, cnt, int64(0))

	assert.Equal(t, len(consumerMsgLst), 0)
}

// 模拟不同的读取方式
func TestReadMode(t *testing.T) {

	msglst := []string{"a", "b", "c", "d", "e", "f", "g"}
	log := blog.BuildWithDefaultOption()
	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	ch1msg := []string{"a", "b", "c", "d", "e", "f", "g"}
	ch2msg := []string{"c", "d", "e", "f", "g"}

	redisps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	ctx := context.TODO()

	topic := redisps.GetTopic("test.topic.readmode")
	defer topic.Close()

	topic.Pub(ctx, &meta.Message{
		Body: []byte(msglst[0]),
	})
	msglst = msglst[1:]
	topic.Pub(ctx, &meta.Message{
		Body: []byte(msglst[0]),
	})
	msglst = msglst[1:]

	ch1, _ := topic.Sub(ctx, "channel.1", WithReadMode(ReadModeBeginning))
	defer ch1.Close()
	ch2, _ := topic.Sub(ctx, "channel.2", WithReadMode(ReadModeLatest))
	defer ch2.Close()

	ch1.Arrived(func(msg *meta.Message) error {
		assert.Equal(t, msg.Body, []byte(ch1msg[0]))
		ch1msg = ch1msg[1:]
		return nil
	})
	ch2.Arrived(func(m *meta.Message) error {
		assert.Equal(t, m.Body, []byte(ch2msg[0]))
		ch2msg = ch2msg[1:]
		return nil
	})

	for _, v := range msglst {
		topic.Pub(ctx, &meta.Message{Body: []byte(v)})
	}

	for {
		<-time.After(time.Second * 2)
		assert.Equal(t, len(ch1msg), 0)
		assert.Equal(t, len(ch2msg), 0)
		break
	}
}

// 模拟多个消费者 - 广播
func TestMultiBroadcastConsumer(t *testing.T) {
	msglst := []string{"a", "b", "c", "d", "e", "f", "g"}
	c1 := make([]string, len(msglst))
	c2 := make([]string, len(msglst))

	copy(c1, msglst)
	copy(c2, msglst)

	log := blog.BuildWithDefaultOption()
	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	ctx := context.TODO()
	var err error

	topic := redisps.GetTopic("test.topic.multi.broadcast")
	defer topic.Close()

	channel1, err := topic.Sub(ctx, "channel.1")
	assert.Equal(t, err, nil) // redis: nil command
	defer channel1.Close()

	channel2, err := topic.Sub(ctx, "channel.2")
	assert.Equal(t, err, nil) // redis: nil command
	defer channel2.Close()

	channel1.Arrived(func(msg *meta.Message) error {
		assert.Equal(t, msg.Body, []byte(c1[0]))
		c1 = c1[1:]
		return nil
	})
	channel2.Arrived(func(msg *meta.Message) error {
		assert.Equal(t, msg.Body, []byte(c2[0]))
		c2 = c2[1:]
		return nil
	})

	for _, v := range msglst {
		topic.Pub(ctx, &meta.Message{Body: []byte(v)})
	}

	for {
		<-time.After(time.Second * 2)
		assert.Equal(t, len(c1), 0)
		assert.Equal(t, len(c2), 0)
		break
	}

}

// 模拟多个消费者 - 轮询
func TestMultiRoundConsumer(t *testing.T) {
	msglst := []string{"a", "b", "c", "d", "e", "f", "g"}
	var c1, c2 int32

	log := blog.BuildWithDefaultOption()
	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	ctx := context.TODO()
	var err error

	topic := redisps.GetTopic("test.topic.multi.round")
	defer topic.Close()

	channel1, err := topic.Sub(ctx, "channel.1")
	assert.Equal(t, err, nil) // redis: nil command
	defer channel1.Close()

	channel2, err := topic.Sub(ctx, "channel.1")
	assert.Equal(t, err, nil) // redis: nil command
	defer channel2.Close()

	channel1.Arrived(func(msg *meta.Message) error {
		atomic.AddInt32(&c1, 1)
		fmt.Println("channel1 arrived", string(msg.Body))
		return nil
	})
	channel2.Arrived(func(msg *meta.Message) error {
		atomic.AddInt32(&c2, 1)
		fmt.Println("channel2 arrived", string(msg.Body))
		return nil
	})

	for _, v := range msglst {
		topic.Pub(ctx, &meta.Message{Body: []byte(v)})
	}

	for {
		<-time.After(time.Second * 2)
		assert.NotEqual(t, c1, 0)
		assert.NotEqual(t, c2, 0)
		assert.Equal(t, int(c1+c2), len(msglst))
		break
	}

}

// 多线程测试
func TestThreadSafe(t *testing.T) {

	log := blog.BuildWithDefaultOption()
	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)
	var tick int32

	ctx := context.TODO()
	var err error

	topic := redisps.GetTopic("test.topic.thread.safe")
	defer topic.Close()

	channel1, err := topic.Sub(ctx, "channel.1")
	assert.Equal(t, err, nil) // redis: nil command
	defer channel1.Close()

	channel1.Arrived(func(msg *meta.Message) error {
		atomic.AddInt32(&tick, 1)
		return nil
	})

	for i := 0; i < 100; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				topic.Pub(ctx, &meta.Message{Body: []byte("test")})
			}
		}()
	}

	for {
		<-time.After(time.Second * 3)
		assert.Equal(t, tick, int32(1000))
		break
	}
}

// ack 相关
//  ack 是否正确回复
//  ack 是否在异常后能够重试
// 网络断开测试
// redis 宕机测试
// 消费者 宕机测试
func TestAck(t *testing.T) {

}

// 消息堆积测试（大量消息未消费被堆积起来，逻辑是否正常
// 消息积压测试（生产大于消费逻辑是否正常
func BenchmarkPubsub(b *testing.B) {

	log := blog.BuildWithDefaultOption()
	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	redisps := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	ctx := context.TODO()
	var err error

	topic := redisps.GetTopic("test.topic.multi.broadcast")
	defer topic.Close()

	channel1, err := topic.Sub(ctx, "channel.1")
	assert.Equal(b, err, nil) // redis: nil command
	defer channel1.Close()

	channel2, err := topic.Sub(ctx, "channel.2")
	assert.Equal(b, err, nil) // redis: nil command
	defer channel2.Close()

	b.ResetTimer()

	channel1.Arrived(func(msg *meta.Message) error {
		return nil
	})

	channel2.Arrived(func(msg *meta.Message) error {
		return nil
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topic.Pub(ctx, &meta.Message{Body: []byte("test")})
		}
	})
}

func TestOptions(t *testing.T) {

}
