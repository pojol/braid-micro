package balancer

import (
	"errors"
	"strings"

	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/pubsub"
)

var (
	m = make(map[string]Builder)
)

// Register 注册balancer
func Register(b Builder) {
	m[strings.ToLower(b.Name())] = b
}

// GetBuilder 获取balancer构建器
func GetBuilder(name string) Builder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}

// Builder 构建器接口
type Builder interface {
	Build(pubsub pubsub.IPubsub) Balancer
	Name() string
}

// Balancer 均衡器接口
type Balancer interface {
	// 选取
	Pick() (nod discover.Node, err error)
}

var (
	// ErrBalanceEmpty 没有权重节点
	ErrBalanceEmpty = errors.New("weighted node list is empty")
	// ErrUninitialized 未初始化
	ErrUninitialized = errors.New("balancer uninitialized")
)
