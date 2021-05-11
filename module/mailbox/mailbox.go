package mailbox

import (
	"encoding/json"
	"strings"

	"github.com/pojol/braid-go/module/logger"
)

// Builder 构建器接口
type Builder interface {
	Build(serviceName string, logger logger.ILogger) (IMailbox, error)
	Name() string
	AddOption(opt interface{})
}

// Message 消息体
type Message struct {
	Body []byte
}

type ScopeTy int32

const (
	ScopeUndefine ScopeTy = 0 + iota
	ScopeProc
	ScopeCluster
)

// NewMessage 构建消息体
func NewMessage(body interface{}) *Message {

	byt, err := json.Marshal(body)
	if err != nil {
		byt = []byte{}
	}

	return &Message{
		Body: byt,
	}
}

type IChannel interface {
	Put(*Message)

	Arrived() <-chan *Message

	Exit() error
}

type ITopic interface {
	// Pub 向 topic 中发送一条消息
	Pub(*Message) error

	// Sub 向 topic 中添加一个用于消费的 channel
	// 如果在一个 topic 中注册同名的 channel 消息仅会被其中的一个消费
	Sub(name string) IChannel

	Exit() error
}

type IMailbox interface {
	// RegistTopic 注册 topic
	RegistTopic(name string, scope ScopeTy) (ITopic, error)

	// GetTopic 获取 mailbox 中的一个 topic （线程安全
	GetTopic(name string) ITopic
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
