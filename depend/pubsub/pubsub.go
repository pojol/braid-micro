// 接口文件 mailbox 邮箱，主要用于包装 Pub-sub 消息模型
package pubsub

// Message 消息体
type Message struct {
	Body []byte
}

// Handler 消息到达的函数句柄
type Handler func(*Message)

// IChannel 信道，topic的子集
type IChannel interface {
	// Arrived 绑定消息到达的函数句柄
	Arrived(Handler)
}

// ITopic 话题，某类消息的聚合
type ITopic interface {
	// Pub 向 topic 中发送一条消息
	Pub(*Message) error

	// Sub 获取topic中的channel
	//
	// 如果在 topic 中没有该 channel 则创建一个新的 channel 到 topic
	//
	// 如果在 topic 中已有同名的 channel 则获取到该 channel
	// 这个时候如果同时有多个 sub 指向同一个 channel 则代表有多个 consumer 对该 channel 进行消费（随机获得
	Sub(channelName string) IChannel

	// RemoveChannel 删除 topic 中存在的 channel
	RemoveChannel(channelName string) error
}

// IPubsub 发布-订阅，管理集群中的所有 Topic
type IPubsub interface {

	// GetTopic 获取 mailbox 中的一个 topic
	//
	// 如果该 topic 不存在，则创建一个新的 topic
	GetTopic(topicName string) ITopic

	// RemoveTopic 删除 mailbox 中存在的 topic
	RemoveTopic(topicName string) error
}
