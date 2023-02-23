package module

type IModule interface {
	Init() error
	Run()
	Close()

	Name() string
}
