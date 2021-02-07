package zaplogger

import "github.com/pojol/braid/module/logger"

// Parm nsq config
type Parm struct {
	filename string
	lv       logger.Lvl

	// maxFileSize is the maximum size in megabytes of the log file before it gets
	maxFileSize int

	// MaxBackups is the maximum number of old log files to retain.  The default
	maxBackups int

	// maxAge is the maximum number of days to retain old log files based on the timestamp encoded in their filename. Note that a day is defined as 24 hours and may not exactly correspond to calendar days due to daylight savings, leap seconds, etc. The default is not to remove old log files based on age.
	maxAge int
}

// Option config wraps
type Option func(*Parm)

// WithFileName 传入日志文件名
func WithFileName(filename string) Option {
	return func(c *Parm) {
		c.filename = filename
	}
}

// WithLv 传入日志等级
func WithLv(lv logger.Lvl) Option {
	return func(c *Parm) {
		c.lv = lv
	}
}

// WithMaxFileSize 传入单个日志文件最大的大小(M
func WithMaxFileSize(fileSize int) Option {
	return func(c *Parm) {
		c.maxFileSize = fileSize
	}
}

// WithMaxBackups 传入最大的日志备份数量
func WithMaxBackups(backups int) Option {
	return func(c *Parm) {
		c.maxBackups = backups
	}
}

// WithMaxAge 文件最大保存天数
func WithMaxAge(age int) Option {
	return func(c *Parm) {
		c.maxAge = age
	}
}
