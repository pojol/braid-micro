package log

import (
	"errors"
	"os"
	"runtime/debug"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type (
	// Config log config
	Config struct {
		Mode   string
		Path   string
		Suffex string
	}

	// Logger logger struct
	Logger struct {
		Path     string
		gSysLog  *zap.Logger
		gSugared *zap.SugaredLogger
	}
)

// 诊断日志
var (
	logPtr *Logger

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

// New new logger
func New(path string) *Logger {
	logPtr = &Logger{
		Path: path,
	}
	return logPtr
}

// Init init logger
func (l *Logger) Init() error {

	lv := zap.DebugLevel

	l.gSysLog = Newlog("", ".sys", lv)
	l.gSugared = l.gSysLog.Sugar()

	return nil
}

// Close 清理日志，并将缓存中的日志刷新到文件
func (l *Logger) Close() {
	l.gSysLog.Sync()
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

/*
	使用非格式化日志，尽量仅用在调试时。
*/

// Debugf debug级别诊断日志
func Debugf(msg string, args ...interface{}) {
	logPtr.gSugared.Debugf(msg, args...)
}

// Fatalf fatal诊断日志
func Fatalf(msg string, args ...interface{}) {

	logPtr.gSugared.Debugf("stask : %v\n", string(debug.Stack()))
	logPtr.gSugared.Fatalf(msg, args...)
}
