package log

import (
	"errors"
	"os"
	"runtime/debug"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

/*
	/ 非格式化日志
		/ 普通日志+调试日志
	/ 格式化日志
		/ 诊断日志
		/ 行为日志
		/ 。。。
*/

type (
	// Config 日志配置项
	Config struct {
		Mode   string
		Path   string
		Suffex string
	}

	// Logger logger struct
	Logger struct {

		// 普通｜调试日志 (支持非结构化
		normalLog     *zap.Logger
		normalSugared *zap.SugaredLogger

		// 系统｜诊断日志（结构化
		SysLog *zap.Logger
		// 用户行为日志（结构化
		BehaviorLog *zap.Logger
	}
)

const (
	// DebugMode debug级别
	DebugMode = "debug"

	// InfoMode info级别
	InfoMode = "info"
)

var (
	logPtr *Logger

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

// New new logger
func New(cfg Config, opts ...Option) *Logger {
	logPtr = &Logger{}

	lv := zap.DebugLevel
	if cfg.Mode == InfoMode {
		lv = zap.InfoLevel
	}

	logPtr.normalLog = Newlog(cfg.Path, cfg.Suffex, lv)
	logPtr.normalSugared = logPtr.normalLog.Sugar()

	for _, opt := range opts {
		opt(logPtr)
	}

	return logPtr
}

// Close 清理日志，并将缓存中的日志刷新到文件
func (l *Logger) Close() {

	if l.normalLog != nil {
		l.normalLog.Sync()
	}

	if l.SysLog != nil {
		l.SysLog.Sync()
	}

	if l.BehaviorLog != nil {
		l.BehaviorLog.Sync()
	}
}

// Newlog 创建一个新的日志
func Newlog(path string, suffex string, lv zapcore.Level) *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   path + suffex, // 日志文件路径
		MaxSize:    64,            // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,            // 日志文件最多保存多少个备份
		MaxAge:     7,             // 文件最多保存多少天
		Compress:   false,         // 是否压缩
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
	atom := zap.NewAtomicLevelAt(lv)
	var ws zapcore.WriteSyncer

	if lv == zap.DebugLevel {
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

// Fatalf fatal诊断日志
func Fatalf(msg string, args ...interface{}) {
	logPtr.normalSugared.Debugf("stask : %v\n", string(debug.Stack()))
	logPtr.normalSugared.Fatalf(msg, args...)
}
