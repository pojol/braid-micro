// 实现文件 基于 nsq 实现的mailbox
package mailboxnsq

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
)

const (
	// Name mailbox plug-in name
	Name = "MailboxNsq"
)

type nsqMailboxBuilder struct {
	opts []interface{}
}

func newNsqMailbox() mailbox.Builder {
	return &nsqMailboxBuilder{}
}

func (nb *nsqMailboxBuilder) AddOption(opt interface{}) {
	nb.opts = append(nb.opts, opt)
}

func (nb *nsqMailboxBuilder) Name() string {
	return Name
}

func (nb *nsqMailboxBuilder) Build(serviceName string, logger logger.ILogger) (mailbox.IMailbox, error) {
	p := Parm{
		ServiceName:       serviceName,
		nsqLogLv:          nsq.LogLevelWarning,
		ConcurrentHandler: 1,
	}
	for _, opt := range nb.opts {
		opt.(Option)(&p)
	}

	rand.Seed(time.Now().UnixNano())
	if len(p.NsqdAddress) != len(p.NsqdHttpAddress) {
		return nil, fmt.Errorf("parm nsqd len(tcp addr) != len(http addr)")
	}

	nsqm := &nsqMailbox{
		parm:     p,
		log:      logger,
		topicMap: make(map[string]*mailboxTopic),
	}

	return nsqm, nil
}

type nsqMailbox struct {
	parm Parm
	log  logger.ILogger

	sync.RWMutex

	topicMap map[string]*mailboxTopic
}

func (nmb *nsqMailbox) RegistTopic(name string, scope mailbox.ScopeTy) (mailbox.ITopic, error) {

	nmb.Lock()
	t, ok := nmb.topicMap[name]
	if ok {
		nmb.Unlock()
		return t, nil
	}

	t = newTopic(name, scope, nmb)
	nmb.topicMap[name] = t
	nmb.Unlock()
	nmb.log.Infof("Topic %v created", name)

	// start loop
	t.start()

	return t, nil
}

func (nmb *nsqMailbox) GetTopic(name string) mailbox.ITopic {

	nmb.RLock()
	t, ok := nmb.topicMap[name]
	nmb.RUnlock()
	if ok {
		return t
	}

	nt, err := nmb.RegistTopic(name, mailbox.ScopeProc)
	if err != nil {
		panic(err)
	}
	nmb.log.Warnf("Get topic warning %v undefined! register proc topic", name)

	return nt
}

func (nmb *nsqMailbox) RemoveTopic(name string) error {
	nmb.RLock()
	topic, ok := nmb.topicMap[name]
	nmb.RUnlock()

	if !ok {
		return fmt.Errorf("topic %v dose not exist", name)
	}

	nmb.log.Infof("deleting topic %v", name)
	topic.Exit()

	nmb.Lock()
	delete(nmb.topicMap, name)
	nmb.Unlock()

	return nil
}

func init() {
	mailbox.Register(newNsqMailbox())
}
