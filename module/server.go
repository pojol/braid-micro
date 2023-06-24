// rpc-server 模块接口文件
package module

// IServer rpc-server interface
type IServer interface {
	Init() error
	Run()
	Close()
}
