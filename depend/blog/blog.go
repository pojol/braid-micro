package blog

import (
	"errors"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	// Logger logger struct
	Logger struct {

		// 普通｜调试日志 (支持非结构化
		normalLog     *zap.Logger
		normalSugared *zap.SugaredLogger
	}
)

const (
	DebugLevel = zap.DebugLevel
	InfoLevel  = zap.InfoLevel
	WarnLevel  = zap.WarnLevel
	ErrLevel   = zap.ErrorLevel
)

var (
	logPtr *Logger

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert linker config")
)

var (
	defaultConfig = Parm{
		lv:      int(InfoLevel),
		size:    64, // default 64 mb
		backups: 30, // saved 30 backup files
		age:     7,  // default save 7 day
		path:    "",
		suffex:  ".braid",
	}
)

func BuildWithOption(opts ...Option) *Logger {

	logParm := &defaultConfig

	for _, opt := range opts {
		opt(logParm)
	}

	return new(logParm)
}

func new(parm *Parm) *Logger {
	logPtr = &Logger{}

	logPtr.normalLog = newlog(parm)
	logPtr.normalSugared = logPtr.normalLog.Sugar()

	return logPtr
}

// Close 清理日志，并将缓存中的日志刷新到文件
func Close() {

	if logPtr != nil {
		logPtr.normalLog.Sync()
	}

}

// Newlog 创建一个新的日志
func newlog(p *Parm) *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   p.path + p.suffex, // 日志文件路径
		MaxSize:    p.size,            // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: p.backups,         // 日志文件最多保存多少个备份
		MaxAge:     p.age,             // 文件最多保存多少天
		Compress:   false,             // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zapcore.Level(p.lv))
	var ws zapcore.WriteSyncer

	if p.stdout {
		ws = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	} else {
		ws = zapcore.NewMultiWriteSyncer( /*zapcore.AddSync(os.Stdout), */ zapcore.AddSync(&hook))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 编码器配置
		ws,                                    // 打印到控制台和文件
		atom,                                  // 日志级别
	)

	logger := zap.New(core)
	return logger
}

// Debugf debug级别诊断日志
func Debugf(msg string, args ...interface{}) {
	logPtr.normalSugared.Debugf(msg, args...)
}

func Infof(msg string, args ...interface{}) {
	logPtr.normalSugared.Infof(msg, args...)
}

func Warnf(msg string, args ...interface{}) {
	logPtr.normalSugared.Warnf(msg, args...)
}

func Errf(msg string, args ...interface{}) {
	logPtr.normalSugared.Errorf(msg, args...)
}
