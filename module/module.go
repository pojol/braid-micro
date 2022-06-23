package module

// IModule module
type IModule interface {

	// Init 模块的初始化阶段
	// 这个阶段主要职责是：部署和检测支撑本模块的运行依赖等...
	Init() error

	// Run 模块的运行期
	// 这个阶段主要职责是：主要用于提供周期性服务，一般会运行在goroutine中。
	Run()

	// Close 关闭模块
	// 这个阶段主要职责是：关闭本模块，并释放模块中依赖的各种资源。
	Close()
}
