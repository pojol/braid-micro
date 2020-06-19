package discover

import (
	"strings"

	"github.com/pojol/braid/plugin/balancer"
)

// Builder 构建器接口
type Builder interface {
	Build(bg *balancer.Group, cfg interface{}) IDiscover
	Name() string
}

// IDiscover 发现服务 & 注册节点
type IDiscover interface {
	Discover()
	Close()
}

var (
	m = make(map[string]Builder)
)

// Register 注册balancer
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
