package client

import (
	"context"
	"strings"

	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
)

// Builder grpc-client builder
type Builder interface {
	Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (IClient, error)
	Name() string
	AddOption(opt interface{})
}

// IClient rpc-client interface
type IClient interface {
	module.IModule

	//
	Invoke(
		ctx context.Context, target, methon, token string,
		args, reply interface{},
		opts ...interface{}) error
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
