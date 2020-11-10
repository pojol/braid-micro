package server

import (
	"strings"

	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/logger"
)

// Builder 构建器接口
type Builder interface {
	Build(serviceName string, logger logger.ILogger) (IServer, error)
	Name() string
	AddOption(opt interface{})
}

// IServer rpc-server interface
type IServer interface {
	module.IModule

	// ...
	Server() interface{}
}

var (
	m = make(map[string]Builder)
)

// Register 注册linker
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
