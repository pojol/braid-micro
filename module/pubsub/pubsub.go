package pubsub

import (
	"encoding/json"
	"strings"
	"time"
)

// Builder 构建器接口
type Builder interface {
	Build() (IPubsub, error)
	Name() string
	SetCfg(cfg interface{}) error
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

// HandlerFunc 消息处理函数
type HandlerFunc func(message *Message) error

// IConsumer 消费者接口
type IConsumer interface {
	OnArrived(handler HandlerFunc)

	PutMsg(msg *Message)

	Exit()
	IsExited() bool
}

// ISubscriber 订阅接口
type ISubscriber interface {

	// AddShared add shared consumer
	AddShared() IConsumer
	// AddCompetition add competition consumer
	AddCompetition() IConsumer

	GetConsumer(cid string) []IConsumer
}

// IPubsub 异步消息通知
type IPubsub interface {
	// Sub 订阅， 创建一个订阅者，它包含一组消费者
	Sub(topic string) ISubscriber

	// Pub produce a message put in topic。
	Pub(topic string, msg *Message)
}

var (
	m = make(map[string]Builder)
)

// Register 注册linker
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
