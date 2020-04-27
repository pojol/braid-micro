package redis

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/mock"
)

func TestMain(m *testing.M) {

	mock.Init()

	code := m.Run()
	// 清理测试环境

	os.Exit(code)
}

func TestRedis(t *testing.T) {
	c := New()
	c.Init(Config{
		Address:        mock.RedisAddr,
		ReadTimeOut:    time.Millisecond * time.Duration(5000),
		WriteTimeOut:   time.Millisecond * time.Duration(5000),
		ConnectTimeOut: time.Millisecond * time.Duration(2000),
		IdleTimeout:    time.Millisecond * time.Duration(0),
		MaxIdle:        16,
		MaxActive:      128,
	})

	assert.Equal(t, c, Get())
	defer c.Close()

	c.ActiveConnCount()
	c.Run()
	c.Ping()
	c.DBSize()
	c.flushDB()
}

// 722519 ns/op
func BenchmarkHGet(b *testing.B) {
	c := New()
	c.Init(Config{
		Address:        mock.RedisAddr,
		ReadTimeOut:    time.Millisecond * time.Duration(5000),
		WriteTimeOut:   time.Millisecond * time.Duration(5000),
		ConnectTimeOut: time.Millisecond * time.Duration(2000),
		IdleTimeout:    time.Millisecond * time.Duration(0),
		MaxIdle:        16,
		MaxActive:      128,
	})
	defer c.Close()

	c.HSet("benchmark_hget", "benchmark_01", "0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.HGet("benchmark_hget", "benchmark_01")
	}
}
