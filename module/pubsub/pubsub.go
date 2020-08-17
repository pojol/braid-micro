package pubsub

import (
	"strings"
	"time"
)

// Builder 构建器接口
type Builder interface {
	Build() IPubsub
	Name() string
}

// Message 消息体
type Message struct {
	ID        string
	Body      interface{}
	Timestamp int64
}

// NewMessage 构建消息体
func NewMessage(body interface{}) *Message {
	return &Message{
		ID:        "",
		Body:      body,
		Timestamp: time.Now().UnixNano(),
	}
}

// HandlerFunc 消息处理函数
type HandlerFunc func(message *Message) error

// IConsumer 消费者接口
type IConsumer interface {
	AddHandler(handler HandlerFunc)

	PutMsg(msg *Message)

	Exit()
	IsExited() bool
}

// IPubsub 异步消息通知
type IPubsub interface {
	// 订阅
	Sub(topic string) IConsumer

	// 通知
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
