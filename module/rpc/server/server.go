// 接口文件 server 定义 RPC 服务的服务端接口
package server

import (
	"github.com/pojol/braid-go/module"
)

// IServer rpc-server interface
type IServer interface {
	module.IModule

	// Server 获取 rpc 的 server 接口
	Server() interface{}
}
