package log

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

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
