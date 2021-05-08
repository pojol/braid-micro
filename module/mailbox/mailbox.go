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
	Channel(name string, scope ScopeTy) IChannel
	Exit() error

	Pub(*Message) error
}

type IMailbox interface {
	Topic(name string) ITopic
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
