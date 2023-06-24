// 服务发现 模块接口文件
package module

// IDiscover discover interface
type IDiscover interface {
	Init() error
	Run()
	Close()
}
