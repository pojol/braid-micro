// 接口文件 server 定义 RPC 服务的服务端接口
package server

// IServer rpc-server interface
type IServer interface {

	// Server 获取 rpc 的 server 接口
	Server() interface{}
}
