package mailboxnsq

import (
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
		ServiceName: serviceName,
		nsqLogLv:    nsq.LogLevelWarning,
	}
	for _, opt := range nb.opts {
		opt.(Option)(&p)
	}

	rand.Seed(time.Now().UnixNano())

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

func (nmb *nsqMailbox) Topic(name string) mailbox.ITopic {

	nmb.RLock()
	t, ok := nmb.topicMap[name]
	nmb.RUnlock()
	if ok {
		return t
	}

	nmb.Lock()
	t, ok = nmb.topicMap[name]
	if ok {
		nmb.Unlock()
		return t
	}

	t = newTopic(name, nmb)
	nmb.topicMap[name] = t
	nmb.Unlock()
	nmb.log.Infof("Topic %v created", name)

	// start loop
	t.start()

	return t
}

func init() {
	mailbox.Register(newNsqMailbox())
}
