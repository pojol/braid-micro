// 链路缓存 模块接口文件
package module

import (
	"context"

	"github.com/pojol/braid-go/module/meta"
)

// ILinkCache 链路缓存，主要用于维护 token 和多个相关联的服务进程之间的关系
//
//	+---parent----------+
//	|                   |
//	|    +--child----+  |
//	|    |           |  |
//	|    | token ... |  |
//	|    |           |  |
//	|    +-----------+  |
//	+-------------------+
type ILinkCache interface {
	Init() error
	Run()
	Close()

	// Target 通过服务名，获取 token 指向的目标服务器地址信息
	Target(ctx context.Context, token string, serviceName string) (targetAddr string, err error)

	// Link 将 token 和目标服务器连接信息写入到缓存中
	Link(ctx context.Context, token string, target meta.Node) error

	// Unlink 将 token 和目标服务器连接信息，解除绑定关系
	Unlink(ctx context.Context, token string) error
}
