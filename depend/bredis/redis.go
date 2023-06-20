package bredis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	defaultConnPoolConfig = &redis.Options{}

	client *redis.Client

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert linker config")
)

func BuildWithOption(opt *redis.Options) *redis.Client {
	return _newRedisClient(opt)
}

// InitWithDefault 基于默认配置进行初始化
func BuildWithDefault() *redis.Client {
	return _newRedisClient(defaultConnPoolConfig)
}

func _newRedisClient(opt *redis.Options) *redis.Client {

	client = redis.NewClient(opt)

	_, err := client.Ping(context.TODO()).Result()
	if err != nil {
		panic(err)
	}

	return client
}
