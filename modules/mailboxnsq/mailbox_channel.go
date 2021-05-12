package mailboxnsq

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/pojol/braid-go/module/mailbox"
)

type IConsumer interface {
	Arrived() <-chan *mailbox.Message
	Put(msg *mailbox.Message)
	Exit()
}

type mailboxChannel struct {
	sync.RWMutex
	clusterConsumer IConsumer

	msgCh chan *mailbox.Message

	mailbox *nsqMailbox

	Name      string
	TopicName string
	scope     mailbox.ScopeTy
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
				ioutil.ReadAll(resp.Body)
				resp.Body.Close()
			}
		}

		var err error
		c.clusterConsumer, err = newClusterConsumer(topicName, channelName, n)
		if err != nil {
			n.log.Errorf("Channel nsq create consumer err %v", err.Error())
		}
	}

	return c
}

func (c *mailboxChannel) Put(msg *mailbox.Message) {
	select {
	case c.msgCh <- msg:
		//default: 就阻塞在这里吧。。
		//	c.mailbox.log.Errorf("put msg err channel %v is full", c.Name)
	}
}

func (c *mailboxChannel) Arrived() <-chan *mailbox.Message {
	if c.scope == mailbox.ScopeProc {
		return c.msgCh
	} else if c.scope == mailbox.ScopeCluster {
		return c.clusterConsumer.Arrived()
	}

	return nil
}

func (c *mailboxChannel) Exit() error {

	return nil
}
