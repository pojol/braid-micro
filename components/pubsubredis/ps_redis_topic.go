package pubsubredis

import (
	"context"
	"fmt"
	"sync"

	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
)

type redisTopic struct {
	sync.RWMutex

	topic string
	log   *blog.Logger

	ps     *redisPubsub
	client *redis.Client

	channelMap map[string]*psRedisChannel
}

func newTopic(name string, client *redis.Client, mgr *redisPubsub, log *blog.Logger) *redisTopic {

	rt := &redisTopic{
		ps:         mgr,
		topic:      name,
		client:     client,
		log:        log,
		channelMap: make(map[string]*psRedisChannel),
	}

	ctx := context.TODO()

	cnt, _ := rt.client.Exists(ctx, rt.topic).Result()
	if cnt == 0 {
		id, err := rt.client.XAdd(ctx, &redis.XAddArgs{
			Stream: rt.topic,
			Values: []string{"msg", "init"},
		}).Result()

		if err != nil {
			log.Warnf("[braid.pubsub ]Topic %v init failed %v", rt.topic, err)
		}

		rt.client.XDel(ctx, rt.topic, id)
	}

	return rt
}

func (rt *redisTopic) Pub(ctx context.Context, msg *meta.Message) error {

	if msg == nil {
		return fmt.Errorf("can't send empty msg to %v", rt.topic)
	}

	// 这里应该包装下

	_, err := rt.client.XAdd(ctx, &redis.XAddArgs{
		Stream: rt.topic,
		Values: []string{"msg", string(msg.Body)},
	}).Result()

	return err
}

func (rt *redisTopic) Sub(ctx context.Context, channel string, opts ...interface{}) (module.IChannel, error) {
	p := ChannelParm{
		ReadMode: ReadModeLatest,
	}

	for _, opt := range opts {
		copt, ok := opt.(ChannelOption)
		if ok {
			copt(&p)
		}
	}

	rt.Lock()
	c, err := rt.getOrCreateChannel(ctx, channel, p)
	rt.Unlock()

	return c, err
}

func (rt *redisTopic) Close() error {

	ctx := context.TODO()
	groups, err := rt.client.XInfoGroups(ctx, rt.topic).Result()

	if len(groups) == 0 {
		cnt, err := rt.client.XLen(ctx, rt.topic).Result()
		if err == nil && cnt == 0 {
			cleanpipe := rt.client.Pipeline()
			cleanpipe.Del(ctx, rt.topic)
			cleanpipe.SRem(ctx, BraidPubsubTopic, rt.topic)

			_, err = cleanpipe.Exec(ctx)
			if err != nil {
				rt.log.Warnf("[braid.pubsub ]Topic %v clean failed %v", rt.topic, err)
			}
		}

	}

	return err
}

func (rt *redisTopic) getOrCreateChannel(ctx context.Context, name string, p ChannelParm) (module.IChannel, error) {

	//channel, ok := rt.channelMap[name]
	//var err error
	//if !ok {
	channel, err := newChannel(ctx, rt.topic, name, rt.client, rt, p)
	if err != nil {
		return nil, err
	}
	rt.channelMap[name] = channel

	rt.log.Infof("[braid.pubsub ]Topic %v new channel %v", rt.topic, name)
	return channel, nil
	//}

	//return channel, nil
}
