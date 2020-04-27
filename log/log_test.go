package log

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestLog(t *testing.T) {
	l := New()
	l.Init(Config{
		Mode:   "debug",
		Path:   "test",
		Suffex: ".log",
	})
	defer l.Close()

	Debugf("msg", 1)
	SysError("log", "testlog", "test log")
	SysSlow("/v1/login/guest", "xxx", 100, "slow request msg")
	SysRoutingError("login", "routing warning")
	SysWelcome("test", "debug", "?", "welcome")
}

func BenchmarkLog(b *testing.B) {
	exePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		b.Error(err)
	}

	l := New()
	l.Init(Config{
		Mode:   "debug",
		Path:   exePath,
		Suffex: ".benchmark",
	})

	defer l.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logPtr.gSysLog.Info("benchmark",
			zap.String("url", "github.com/lestrrat-go/file-rotatelogs"),
			zap.Int("attempt", 3),
			zap.Duration("backoff", time.Second),
		)
	}
}
