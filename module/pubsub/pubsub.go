package pubsub

import (
	"strings"

	"github.com/pojol/braid/internal/braidsync"
)

// Builder 构建器接口
type Builder interface {
	Build() IPubsub
	Name() string
}

// IPubsub 异步消息通知
type IPubsub interface {
	Sub(topic string) *braidsync.Unbounded
	Pub(topic string, msg interface{})
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

/*
// 内部

ps.Pub("braid_event_discover", discover.Nod{})

balancer.Init() {
	discoverCH = balancer.Sub("braid_event_discover")

	go func() {
		for {
			select {
			case <-discoverCH:
				todo ...
			}
		}
	}()
}

// 跨进程
*/
