package blog

type Parm struct {
	lv       int
	size     int
	backups  int
	age      int
	compress bool
	stdout   bool

	path   string
	suffex string
}

// Option config wraps
type Option func(*Parm)

// WithLevel 日志等级
func WithLevel(loglv int) Option {
	return func(c *Parm) {
		c.lv = loglv
	}
}

// WithMaxSize 日志文件保存的最大 mb
func WithMaxSize(size int) Option {
	return func(c *Parm) {
		c.size = size
	}
}

// WithBackups 最多保存的备份文件数
func WithBackups(backupday int) Option {
	return func(c *Parm) {
		c.backups = backupday
	}
}

// WithMaxAge 最大备份天数
func WithMaxAge(age int) Option {
	return func(c *Parm) {
		c.age = age
	}
}

// WithStdout 开启控制台输出
func WithStdout(stdout bool) Option {
	return func(c *Parm) {

	}
}

// WithCompress 是否压缩
func WithCompress(compress bool) Option {
	return func(c *Parm) {
		c.compress = compress
	}
}
