package linkcache

import (
	"strings"
)

// Builder 构建器接口
type Builder interface {
	Build(serviceName string) (ILinkCache, error)
	Name() string
	AddOption(opt interface{})
}

// ILinkCache The connector is a service that maintains the link relationship between multiple processes and users.
//
// +---parent----------+
// |                   |
// |    +--child----+  |
// |    |           |  |
// |    | token ... |  |
// |    |           |  |
// |    +-----------+  |
// |                   |
// +-------------------+
type ILinkCache interface {
	// Look for existing links from the cache
	Target(child string, token string) (targetAddr string, err error)

	// 将token绑定到nod
	Link(clild string, token string, targetAddr string) error

	// unlink token
	Unlink(token string) error

	// 提供nod中token的数量
	Num(clild string, targetAddr string) (int, error)

	// clean up the service
	Down(clild string, targetAddr string) error
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
