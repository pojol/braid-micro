package redis

import (
	"os"
	"testing"
	"time"

	"github.com/pojol/braid-go/mock"
	"github.com/stretchr/testify/assert"
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

	conn := c.Conn()
	defer conn.Close()

	testkey := "redis_test"
	testhfield := "redis_h_field"

	c.Set(testkey, "test")
	c.Expire(testkey, 1)
	c.SetWithExpire(testkey, "test", 1)
	c.SetEx(testkey, 1, "test")
	c.Get(testkey)
	c.Del(testkey)

	c.HSet(testhfield, testkey, "test")
	c.HGet(testhfield, testkey)
	c.HGetAll(testhfield)
	c.HKeys(testhfield)
	c.HExist(testhfield, testkey)
	c.HLen(testhfield)
	c.HDel(testhfield, testkey)

	c.LPush(testkey, "1")
	c.RPop(testkey)
	c.RPush(testkey, "1")
	c.LLen(testkey)
	c.LRange(testkey, 0, -1)
	c.LRem(testkey, 1, "1")

	c.Keys("*")
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
