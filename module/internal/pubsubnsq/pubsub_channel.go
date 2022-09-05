package pubsubnsq

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/module/pubsub"
)

type pubsubChannel struct {
	sync.RWMutex

	msgCh *UnboundedMsg

	ps       *nsqPubsub
	exitFlag int32

	consumer *nsq.Consumer

	Name      string
	TopicName string
}

type consumerHandler struct {
	channel string
	c       *pubsubChannel
}

func (ch *consumerHandler) HandleMessage(msg *nsq.Message) error {

	ch.c.Put(&pubsub.Message{
		Body: msg.Body,
	})
	return nil
}

func newChannel(topicName, channelName string, n *nsqPubsub) *pubsubChannel {

	c := &pubsubChannel{
		Name:      channelName,
		TopicName: topicName,
		ps:        n,
		msgCh:     NewUnbounded(),
	}

	for _, addr := range n.parm.NsqdHttpAddress {
		url := fmt.Sprintf("http://%s/channel/create?topic=%s&channel=%s",
			addr,
			topicName,
			channelName,
		)

		req, _ := http.NewRequest("POST", url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			blog.Warnf("%v request err %v", url, err.Error())
		}

		if resp != nil {
			if resp.StatusCode != http.StatusOK {
				blog.Warnf("nsqd create channel request status err %v", resp.StatusCode)
			}

			ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}

	cfg := nsq.NewConfig()
	cfg.MaxInFlight = len(n.parm.NsqdHttpAddress)
	nsqConsumer, err := nsq.NewConsumer(topicName, channelName, cfg)
	if err != nil {
		blog.Errf("channel %v nsq.NewConsumer err %v", channelName, err)
		return nil
	}
	nsqConsumer.SetLoggerLevel(n.parm.NsqLogLv)

	nsqConsumer.AddConcurrentHandlers(&consumerHandler{
		c:       c,
		channel: channelName,
	}, int(n.parm.ConcurrentHandler))

	if len(n.parm.LookupdAddress) == 0 { // 不推荐的处理方式
		err = nsqConsumer.ConnectToNSQDs(n.parm.NsqdAddress)
		if err != nil {
			blog.Errf("channel %v nsq.ConnectToNSQDs err %v", channelName, err)
			return nil
		}
	} else {
		err = nsqConsumer.ConnectToNSQLookupds(n.parm.LookupdAddress)
		if err != nil {
			blog.Errf("channel %v nsq.ConnectToNSQLookupds err %v", channelName, err)
			return nil
		}
	}

	c.consumer = nsqConsumer
	//blog.Infof("Cluster consumer %v created", channelName)

	return c
}

func (c *pubsubChannel) Put(msg *pubsub.Message) {

	if atomic.LoadInt32(&c.exitFlag) == 1 {
		blog.Warnf("cannot write to the exiting channel %v", c.Name)
		return
	}

	c.msgCh.Put(msg)
}

func (c *pubsubChannel) addHandlers(handler pubsub.Handler) {
	go func() {
		for {
			m, ok := <-c.msgCh.Get()
			if !ok {
				goto EXT
			}
			c.msgCh.Load()

			handler(m)
		}
	EXT:
		blog.Infof("channel %v stopping handler", c.Name)
	}()
}

func (c *pubsubChannel) Arrived(handler pubsub.Handler) {
	c.addHandlers(handler)
}

func (c *pubsubChannel) Exit() error {
	if !atomic.CompareAndSwapInt32(&c.exitFlag, 0, 1) {
		return errors.New("exiting")
	}

	//blog.Infof("channel %v exiting", c.Name)

	c.consumer.Stop()

	return nil
}
