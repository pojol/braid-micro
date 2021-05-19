// 接口文件 logger 日志接口
package logger

import "strings"

// Lvl log lv
type Lvl uint8

// log lv
const (
	DEBUG Lvl = iota + 1
	INFO
	WARN
	ERROR
)

// Builder grpc-client builder
type Builder interface {
	Build() (ILogger, error)
	Name() string
	AddOption(opt interface{})
}

// ILogger logger
type ILogger interface {
	Debug(i ...interface{})
	Debugf(format string, args ...interface{})

	Info(i ...interface{})
	Infof(format string, args ...interface{})

	Warn(i ...interface{})
	Warnf(format string, args ...interface{})

	Error(i ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(i ...interface{})
	Fatalf(format string, args ...interface{})
}

var (
	m = make(map[string]Builder)
)

// Register 注册balancer
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
