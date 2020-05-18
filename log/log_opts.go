package log

import "go.uber.org/zap"

// Option config wraps
type Option func(*Logger)

// WithSysLog 添加系统日志
func WithSysLog(cfg Config) Option {

	return func(log *Logger) {
		lv := zap.DebugLevel
		if cfg.Mode == InfoMode {
			lv = zap.InfoLevel
		}

		log.SysLog = Newlog(cfg.Path, cfg.Suffex, lv)
	}

}

// WithBehavior 添加行为日志
func WithBehavior(cfg Config) Option {
	return func(log *Logger) {
		lv := zap.DebugLevel
		if cfg.Mode == InfoMode {
			lv = zap.InfoLevel
		}

		log.BehaviorLog = Newlog(cfg.Path, cfg.Suffex, lv)
	}
}
