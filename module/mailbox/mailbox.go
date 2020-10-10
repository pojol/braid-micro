package mailbox

import (
	"encoding/json"
	"strings"
	"time"
)

// Builder 构建器接口
type Builder interface {
	Build(serviceName string) (IMailbox, error)
	Name() string
	AddOption(opt interface{})
}

// Message 消息体
type Message struct {
	ID        string
	Body      []byte
	Timestamp int64
}

// NewMessage 构建消息体
func NewMessage(body interface{}) *Message {

	byt, err := json.Marshal(body)
	if err != nil {
		byt = []byte{}
	}

	return &Message{
		ID:        "",
		Body:      byt,
		Timestamp: time.Now().UnixNano(),
	}
}

// HandlerFunc msg handler
type HandlerFunc func(message *Message) error

// IConsumer consumer
type IConsumer interface {
	OnArrived(handler HandlerFunc)

	PutMsg(msg *Message) error

	Exit()
	IsExited() bool
}

// ISubscriber subscriber
type ISubscriber interface {
	AddShared() (IConsumer, error)
	AddCompetition() (IConsumer, error)
}

// IMailbox mailbox
type IMailbox interface {
	ProcPub(topic string, msg *Message)
	ProcSub(topic string) ISubscriber

	ClusterPub(topic string, msg *Message)
	ClusterSub(topic string) ISubscriber
}

var (
	m = make(map[string]Builder)
)

// Register 注册
func Register(b Builder) {
	m[strings.ToLower(b.Name())] = b
}

// GetBuilder 获取构建器
func GetBuilder(name string) Builder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}
