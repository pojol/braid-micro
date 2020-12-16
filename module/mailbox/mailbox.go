package mailbox

import (
	"encoding/json"
	"strings"
	"sync"
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

// MessageBuffer 仅用于 mailbox 的 unbounded
type MessageBuffer struct {
	c       chan *Message
	backlog []*Message

	sync.Mutex
}

// NewMessageBuffer 构建 unbounded message buffer
func NewMessageBuffer() *MessageBuffer {
	return &MessageBuffer{
		c: make(chan *Message, 1),
	}
}

// Put put msg
func (mbuffer *MessageBuffer) Put(msg *Message) {
	mbuffer.Lock()

	if len(mbuffer.backlog) == 0 {
		select {
		case mbuffer.c <- msg:
			mbuffer.Unlock()
			return
		default:
		}
	}

	mbuffer.backlog = append(mbuffer.backlog, msg)
	mbuffer.Unlock()
}

// Load 将积压队列中的头部数据提取到channel，并将队列整体前移一位。
func (mbuffer *MessageBuffer) Load() {
	mbuffer.Lock()

	if len(mbuffer.backlog) > 0 {
		select {
		case mbuffer.c <- mbuffer.backlog[0]:
			mbuffer.backlog[0] = nil
			mbuffer.backlog = mbuffer.backlog[1:]
		default:
		}
	}

	mbuffer.Unlock()
}

// Get 获取 read channel
func (mbuffer *MessageBuffer) Get() <-chan *Message {
	return mbuffer.c
}

// HandlerFunc msg handler
type HandlerFunc func(message Message) error

// IConsumer consumer
type IConsumer interface {
	OnArrived(handler HandlerFunc) error

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
