package link

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/cache/redis"
	"github.com/pojol/braid/mock"
)

func TestTarget(t *testing.T) {

	mock.Init()

	r := redis.New()
	r.Init(redis.Config{
		Address:        mock.RedisAddr,
		ReadTimeOut:    time.Millisecond * time.Duration(5000),
		WriteTimeOut:   time.Millisecond * time.Duration(5000),
		ConnectTimeOut: time.Millisecond * time.Duration(2000),
		IdleTimeout:    time.Millisecond * time.Duration(0),
		MaxIdle:        16,
		MaxActive:      128,
	})

	l := New()
	l.Init(Config{})

	err := l.Offline("address")
	assert.Equal(t, err, nil)

	err = l.Link(context.Background(), "test_link", "address")
	assert.Equal(t, err, nil)

	addr, err := l.Target(context.Background(), "test_link")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "address")

	num, err := l.Num("address")
	assert.Equal(t, num, 1)

	err = l.Offline("address")
	assert.Equal(t, err, nil)

	assert.Equal(t, Get(), l)
	l.Unlink("")
	l.Run()
	l.Close()

}
