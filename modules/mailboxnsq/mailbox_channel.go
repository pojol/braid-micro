package mailboxnsq

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/module/mailbox"
)

type mailboxChannel struct {
	sync.RWMutex

	msgCh chan *mailbox.Message

	mailbox  *nsqMailbox
	exitFlag int32

	consumer *nsq.Consumer

	Name      string
	TopicName string
	scope     mailbox.ScopeTy
}

type consumerHandler struct {
	channel string
	c       *mailboxChannel
}

func (ch *consumerHandler) HandleMessage(msg *nsq.Message) error {

	ch.c.Put(&mailbox.Message{
		Body: msg.Body,
	})
	return nil
}

func newChannel(topicName, channelName string, scope mailbox.ScopeTy, n *nsqMailbox) *mailboxChannel {

	c := &mailboxChannel{
		Name:      channelName,
		TopicName: topicName,
		scope:     scope,
		mailbox:   n,
		msgCh:     make(chan *mailbox.Message, 10000),
	}

	if scope == mailbox.ScopeCluster {

		for _, addr := range n.parm.NsqdHttpAddress {
			url := fmt.Sprintf("http://%s/channel/create?topic=%s&channel=%s",
				addr,
				topicName,
				channelName,
			)

			req, _ := http.NewRequest("POST", url, nil)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				n.log.Warnf("%v request err %v", url, err.Error())
			}

			if resp != nil {
				if resp.StatusCode != http.StatusOK {
					n.log.Warnf("nsqd create channel request status err %v", resp.StatusCode)
				}

				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
			}
		}

		nsqConsumer, err := nsq.NewConsumer(topicName, channelName, nsq.NewConfig())
		if err != nil {
			n.log.Errorf("channel %v nsq.NewConsumer err %v", channelName, err)
			return nil
		}
		nsqConsumer.SetLoggerLevel(n.parm.nsqLogLv)

		nsqConsumer.AddConcurrentHandlers(&consumerHandler{
			c:       c,
			channel: channelName,
		}, int(n.parm.ConcurrentHandler))

		if len(n.parm.LookupdAddress) == 0 { // 不推荐的处理方式
			err = nsqConsumer.ConnectToNSQDs(n.parm.NsqdAddress)
			if err != nil {
				n.log.Errorf("channel %v nsq.ConnectToNSQDs err %v", channelName, err)
				return nil
			}
		} else {
			err = nsqConsumer.ConnectToNSQLookupds(n.parm.LookupdAddress)
			if err != nil {
				n.log.Errorf("channel %v nsq.ConnectToNSQLookupds err %v", channelName, err)
				return nil
			}
		}

		c.consumer = nsqConsumer
		n.log.Infof("Cluster consumer %v created", channelName)
	}

	return c
}

func (c *mailboxChannel) Put(msg *mailbox.Message) {

	if atomic.LoadInt32(&c.exitFlag) == 1 {
		c.mailbox.log.Warnf("cannot write to the exiting channel %v", c.Name)
		return
	}

	select {
	case c.msgCh <- msg:
	default: // channel 被写满，且没有来得及消费掉
		c.mailbox.log.Errorf("put msg err channel %v is full", c.Name)
	}
}

func (c *mailboxChannel) addHandlers(handler mailbox.Handler) {
	go func() {
		for {
			m, ok := <-c.msgCh
			if !ok {
				goto EXT
			}

			handler(m)
		}
	EXT:
		c.mailbox.log.Infof("channel %v stopping handler", c.Name)
	}()
}

func (c *mailboxChannel) Arrived(handler mailbox.Handler) {
	c.addHandlers(handler)
}

func (c *mailboxChannel) Exit() error {
	if !atomic.CompareAndSwapInt32(&c.exitFlag, 0, 1) {
		return errors.New("exiting")
	}

	c.mailbox.log.Infof("channel %v exiting", c.Name)

	if c.scope == mailbox.ScopeCluster {
		c.consumer.Stop()
	}
	close(c.msgCh)

	return nil
}
