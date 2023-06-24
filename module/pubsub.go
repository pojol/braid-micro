// 消息发布订阅 模块接口文件
package module

import (
	"context"

	"github.com/pojol/braid-go/module/meta"
)

// Handler 消息到达的函数句柄
type Handler func(*meta.Message) error

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
	Pub(ctx context.Context, msg *meta.Message) error

	// Sub 向 topic 订阅消息
	//  注: 消费者只能保证消息必定被消费一次，但不保证消费只会被消费一次
	Sub(ctx context.Context, channelName string, opts ...interface{}) (IChannel, error)

	// Close 关闭 topic
	Close() error
}

// IPubsub 发布-订阅，管理集群中的所有 Topic
type IPubsub interface {
	// Topic 获取一个 topic，如果没有则创建一个新的
	GetTopic(name string) ITopic

	// Info 输出topic的信息
	Info()
}
