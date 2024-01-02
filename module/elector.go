// 选举 模块接口文件
package module

// IElector election interface
type IElector interface {
	Init() error
	Run()
	Close()
}
