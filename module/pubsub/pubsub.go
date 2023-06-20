// 接口文件 pubsub
package pubsub

import "context"

// Message 消息体
type Message struct {
	Body []byte
}

// Handler 消息到达的函数句柄
type Handler func(*Message)

// IChannel 信道，topic的消费队列
type IChannel interface {
	// Arrived 绑定消息到达的函数句柄
	Arrived(Handler)

	// 关闭 channel 在topic 中的订阅，并退出
	Close() error
}

// ITopic 话题，消息对象
type ITopic interface {
	// Pub 向 topic 发送一条消息
	Pub(ctx context.Context, msg *Message) error

	// Sub 向 topic 订阅消息
	//  channelName 信道名称，如果订阅的信道已经存在（同名）则每个消费端随机消费消息
	//  如果不存在，则创建一个新的信道，每个信道都可以消费 topic 的全部消息
	Sub(ctx context.Context, channelName string) IChannel

	// Close 关闭 topic
	Close() error
}

// IPubsub 发布-订阅，管理集群中的所有 Topic
type IPubsub interface {
	Topic(name string) ITopic
}
