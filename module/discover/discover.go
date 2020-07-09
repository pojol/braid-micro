package discover

import (
	"strings"
)

// Builder 构建器接口
type Builder interface {
	Build() IDiscover
	Name() string
	SetCfg(cfg interface{}) error
}

// Nod 发现节点结构
type Nod struct {
	ID      string
	Name    string
	Address string
	Meta    interface{}
}

// Callback 发现回掉
type Callback func(nod Nod)

// IDiscover 发现服务 & 注册节点
type IDiscover interface {
	Discover()
	Close()
}

var (
	m = make(map[string]Builder)

	discov IDiscover
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
