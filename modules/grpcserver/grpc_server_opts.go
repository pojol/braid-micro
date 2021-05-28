package grpcserver

import (
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
)

// Parm Service 配置
type Parm struct {
	ListenAddr string

	openRecover   bool
	recoverHandle grpc_recovery.RecoveryHandlerFunc
}

// Option config wraps
type Option func(*Parm)

// WithListen 服务器侦听地址配置
func WithListen(address string) Option {
	return func(c *Parm) {
		c.ListenAddr = address
	}
}

// WithRecover 设置是否开启recover，当开启时 内部的 panic 将转换为 err
func WithRecover(f grpc_recovery.RecoveryHandlerFunc) Option {
	return func(c *Parm) {
		c.openRecover = true
		c.recoverHandle = f
	}
}
