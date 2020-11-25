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

const (
	// Proc 发布一个进程内消息
	Proc = "mailbox.proc"

	// Cluster 发布一个集群消息
	Cluster = "mailbox.cluster"
)

const (
	// Undecided 暂未决定的
	Undecided = "mailbox.undecided"

	// Competition 竞争型的信道（只被消费一次
	Competition = "mailbox.competition"

	// Shared 共享型的信道, 消息副本会传递到多个消费者
	Shared = "mailbox.shared"
)

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
type HandlerFunc func(message Message) error

// IConsumer consumer
type IConsumer interface {
	OnArrived(handler HandlerFunc)

	PutMsg(msg *Message) error

	Exit()
	IsExited() bool
}

// ISubscriber 订阅者
type ISubscriber interface {
	Shared() (IConsumer, error)
	Competition() (IConsumer, error)
}

// IMailbox mailbox
type IMailbox interface {
	Pub(scope string, topic string, msg *Message)

	Sub(scope string, topic string) ISubscriber
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
