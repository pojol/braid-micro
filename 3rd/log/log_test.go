package log

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestLog(t *testing.T) {
	l := New(Config{
		Mode:   DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, WithSys(Config{
		Mode:   DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	Debugf("msg", 1)
	SysError("log", "testlog", "test log")
	SysSlow("/v1/login/guest", "xxx", 100, "slow request msg")
	SysRoutingError("login", "routing warning")
	SysWelcome("test", "debug", "?", "welcome")
	SysCompose([]string{"11"}, "compose")
	SysElection("master", "xx10")
}

func TestOpts(t *testing.T) {

	var tests = []struct {
		Mod string
	}{
		{DebugMode},
		{InfoMode},
	}

	for _, v := range tests {
		dl := New(Config{
			Mode:   v.Mod,
			Path:   "testNormal",
			Suffex: ".log",
		}, WithSys(Config{
			Mode:   v.Mod,
			Path:   "testSys",
			Suffex: ".sys",
		}), WithBehavior(Config{
			Mode:   v.Mod,
			Path:   "testBehavior",
			Suffex: ".Behavior",
		}))
		defer dl.Close()
	}

}

func BenchmarkLog(b *testing.B) {
	exePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		b.Error(err)
	}
	l := New(Config{
		Mode:   DebugMode,
		Path:   exePath + "/testNormal",
		Suffex: ".log",
	}, WithSys(Config{
		Mode:   DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logPtr.SysLog.Info("benchmark",
			zap.String("url", "github.com/lestrrat-go/file-rotatelogs"),
			zap.Int("attempt", 3),
			zap.Duration("backoff", time.Second),
		)
	}
}
