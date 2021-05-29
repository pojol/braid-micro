// 接口文件 logger 日志接口
package logger

// Lvl log lv
type Lvl uint8

// log lv
const (
	DEBUG Lvl = iota + 1
	INFO
	WARN
	ERROR
)

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
