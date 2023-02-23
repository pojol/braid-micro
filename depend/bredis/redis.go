package bredis

import (
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
	client = redis.NewClient(opt)

	return client
}

// InitWithDefault 基于默认配置进行初始化
func BuildWithDefault() *redis.Client {
	client = redis.NewClient(defaultConnPoolConfig)

	return client
}
