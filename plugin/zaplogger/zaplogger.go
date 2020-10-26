package zaplogger

import (
	"os"

	"github.com/pojol/braid/module/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Name log name
	Name = "ZapLogger"
)

type zaplogBuilder struct {
}

func newZapLogger() logger.Builder {
	return &zaplogBuilder{}
}

func (*zaplogBuilder) Name() string {
	return Name
}

func (*zaplogBuilder) Build(lv logger.Lvl) (logger.ILogger, error) {

	var atom zap.AtomicLevel
	var ws zapcore.WriteSyncer

	if lv == logger.DEBUG {
		atom = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else if lv == logger.INFO {
		atom = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else if lv == logger.WARN {
		atom = zap.NewAtomicLevelAt(zap.WarnLevel)
	} else {
		atom = zap.NewAtomicLevelAt(zap.ErrorLevel)
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

	hook := lumberjack.Logger{
		Filename:   "log.braid", // 日志文件路径
		MaxSize:    64,          // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,          // 日志文件最多保存多少个备份
		MaxAge:     7,           // 文件最多保存多少天
		Compress:   false,       // 是否压缩
	}

	if lv == logger.ERROR {
		ws = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	} else {
		ws = zapcore.NewMultiWriteSyncer(zapcore.AddSync(&hook))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 编码器配置
		ws,                                    // 打印到控制台和文件
		atom,                                  // 日志级别
	)

	return zap.New(core).Sugar(), nil
}

func init() {
	logger.Register(newZapLogger())
}
