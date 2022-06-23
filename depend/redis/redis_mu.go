package redis

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	// Expiry 默认2秒的超时时间，当到达超时时强行释放锁。
	Expiry = 3 * time.Second

	// Tries 如果获取锁失败，可重试的次数
	Tries = 4

	// Delay 重新获得锁的间隔(毫秒
	Delay = 700
)

// Mutex 分布式token锁
type Mutex struct {
	Token string
	value string
}

var (
	// ErrRedisMtxAcquire 申请分布式锁失败
	ErrRedisMtxAcquire = errors.New("failed to acquire lock")
	// ErrRedisMtxInvaildToken 错误的传入token
	ErrRedisMtxInvaildToken = errors.New("acquire token is empty")
)

// Lock 分布式锁
func (m *Mutex) Lock(from string) error {

	if m.Token == "" {
		return ErrRedisMtxInvaildToken
	}

	b := make([]byte, 16)
	rand.Read(b)

	value := base64.StdEncoding.EncodeToString(b)
	for i := 0; i < Tries; i++ {

		conn := client.pool.Get()
		reply, err := redis.String(conn.Do("set", m.Token, value, "nx", "px", int(Expiry/time.Millisecond)))
		conn.Close()

		if err == nil && reply == "OK" {
			m.value = value
			return nil
		}

		time.Sleep(time.Duration(time.Millisecond * Delay))
	}

	return ErrRedisMtxAcquire
}

// Unlock 释放锁
func (m *Mutex) Unlock() bool {

	value := m.value
	if value == "" {
		return false
	}

	m.value = ""

	conn := client.pool.Get()
	defer conn.Close()
	status, err := GetDelScript.Do(conn, m.Token, value)

	return status != 0 && err == nil
}

// GetDelScript 将get 和del整合成原子操作
var GetDelScript = redis.NewScript(1, `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)
