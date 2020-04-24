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
