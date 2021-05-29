// 实现文件 zaplogger 基于 zap 实现的logger
package zaplogger

import (
	"os"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Name log name
	Name = "ZapLogger"
)

type zaplogBuilder struct {
	opts []interface{}
}

func newZapLogger() module.IBuilder {
	return &zaplogBuilder{}
}

func (zb *zaplogBuilder) Name() string {
	return Name
}

func (zb *zaplogBuilder) Type() module.ModuleType {
	return module.Logger
}

func (zb *zaplogBuilder) AddModuleOption(opt interface{}) {
	zb.opts = append(zb.opts, opt)
}

func (zb *zaplogBuilder) Build(name string, buildOpts ...interface{}) interface{} {

	p := Parm{
		filename:    "log.braid",
		lv:          logger.DEBUG,
		maxFileSize: 64,
		maxBackups:  30,
		maxAge:      7,
	}

	for _, opt := range zb.opts {
		opt.(Option)(&p)
	}

	var atom zap.AtomicLevel
	var ws zapcore.WriteSyncer

	if p.lv == logger.DEBUG {
		atom = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else if p.lv == logger.INFO {
		atom = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else if p.lv == logger.WARN {
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
		Filename:   p.filename,    // 日志文件路径
		MaxSize:    p.maxFileSize, // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: p.maxBackups,  // 日志文件最多保存多少个备份
		MaxAge:     p.maxAge,      // 文件最多保存多少天
		Compress:   false,         // 是否压缩
	}

	ws = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 编码器配置
		ws,                                    // 打印到控制台和文件
		atom,                                  // 日志级别
	)

	return zap.New(core).Sugar()
}

func init() {
	module.Register(newZapLogger())
}
