package pubsubredis

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/internal/buffer"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
)

type psRedisChannel struct {
	topic    string
	channel  string // stream group
	consumer string // group consumer
	client   *redis.Client
	log      *blog.Logger

	exitFlag int32
	msgCh    *buffer.UnboundedMsg
}

func newChannel(ctx context.Context, topic, channel string, client *redis.Client, rt *redisTopic, p ChannelParm) (*psRedisChannel, error) {

	c := &psRedisChannel{
		topic:    topic,
		channel:  channel,
		consumer: uuid.New().String(),

		log: rt.log,

		msgCh: buffer.NewUUnboundedMsg(),

		client: client,
	}

	// 从头部开始消费，还是从最新的消息开始 (默认从尾部开始进行消费，只处理新消息
	_, err := client.XGroupCreate(ctx, topic, c.channel, p.ReadMode).Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return nil, err
	}
	c.loop()

	return c, nil
}

func (c *psRedisChannel) loop() {
	go func() {
		for {
			msgs := c.client.XReadGroup(context.TODO(), &redis.XReadGroupArgs{
				Group:    c.channel,
				Consumer: c.consumer,
				Streams:  []string{c.topic, ">"},
				Block:    100 * time.Millisecond,
				Count:    10,
			}).Val()

			for _, v := range msgs {
				for _, msg := range v.Messages {

					val := msg.Values["msg"].(string)

					if atomic.LoadInt32(&c.exitFlag) == 1 {
						//c.log.Warnf("cannot write to the exiting channel %v", c.Name)
						return
					}
					c.msgCh.Put(meta.CreateMessage(msg.ID, []byte(val)))
				}
			}

		}
	}()
}

func (c *psRedisChannel) addHandlers(handler module.Handler) {
	go func() {
		for {
			m, ok := <-c.msgCh.Get()
			if !ok {
				goto EXT
			}
			c.msgCh.Load()

			pipe := c.client.Pipeline()

			err := handler(m)
			if err == nil {
				pipe.XAck(context.TODO(), c.topic, c.channel, m.ID())
				pipe.XDel(context.TODO(), c.topic, m.ID())
			}

			_, err = pipe.Exec(context.TODO())
			if err != nil {
				c.log.Warnf("topic %v channel %v id %v pipeline failed: %v", c.topic, c.channel, m.ID(), err)
			}
		}
	EXT:
		//c.log.Infof("channel %v stopping handler", c.Name)
	}()
}

func (c *psRedisChannel) Arrived(handler module.Handler) {
	c.addHandlers(handler)
}

func (c *psRedisChannel) Close() error {

	_, err := c.client.XGroupDelConsumer(context.TODO(), c.topic, c.channel, c.consumer).Result()
	if err != nil {
		c.log.Warnf("braid.pubsub topic %v channel %v redis channel del consumer err %v", c.topic, c.channel, err.Error())
		return err
	}

	consumers, err := c.client.XInfoConsumers(context.TODO(), c.topic, c.channel).Result()
	if err != nil {
		c.log.Warnf("braid.pubsub topic %v channel %v redis channel info consumers err %v", c.topic, c.channel, err.Error())
		return err
	}

	if len(consumers) == 0 {
		_, err := c.client.XGroupDestroy(context.TODO(), c.topic, c.channel).Result()
		if err != nil {
			c.log.Warnf("braid.pubsub topic %v channel %v redis channel destory err %v", c.topic, c.channel, err.Error())
		}
	}

	return err
}
