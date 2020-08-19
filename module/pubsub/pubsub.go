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
	OnArrived(handler HandlerFunc)

	PutMsg(msg *Message)

	Exit()
	IsExited() bool
}

// ISubscriber 订阅接口
type ISubscriber interface {
	// 添加消费者，竞争消费
	AddConsumer() IConsumer
	// 添加消费者，共同消费
	AppendConsumer() IConsumer

	PutMsg(msg *Message)
}

// IPubsub 异步消息通知
type IPubsub interface {
	// 订阅， 创建一个消费者组
	// 1 topic : 1 consumer
	// sub1 := pubsub.Sub("discover_add")
	// sub.AddConsumer().OnArrived( func(msg *pubsub.Message) error { return nil } )
	// sub2 := pubsub.Sub("discover_rmv")
	// sub2.AddConsumer().OnArrived()
	//
	// 1 topic : N consumer (一个topic 被多个consumer 竞争消费
	// sub1 := pubsub.Sub("discover_add").AddConsumer().OnArrived()
	// sub2 := pubsub.Sub("discover_add").AddConsumer().OnArrived()
	//
	// 1 topic : N consumer (一个topic 拥有多个consumer 共同消费
	// sub := pubsub.Sub("discover_add")
	// sub.AddConsumer().OnArrived()
	// sub.AppendConsumer().OnArrived()
	Sub(topic string) ISubscriber

	// 生产一条 message 投送到 topic。
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
