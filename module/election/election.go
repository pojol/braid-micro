package election

import (
	"strings"
)

// Builder 构建器接口
type Builder interface {
	Build(cfg interface{}) (IElection, error)
	Name() string
}

// IElection 选举器需要提供的接口
type IElection interface {
	// 
	IsMaster() bool

	Run()
	Close()
}

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
