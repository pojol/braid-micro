package link

import (
	"errors"

	"github.com/pojol/braid/cache/redis"
)

type (
	// Linker 链接器
	Linker struct {
	}

	// Config 配置项目
	Config struct {
	}
)

const (
	// RedisTokenAddressHash 保存token的绑定地址
	// 这里存储着玩家令牌和服务之间的指向关系，当服务产生故障无法正常服务时，应该将令牌和此服务解除绑定。
	// "token1" -> 127.0.0.1:8001
	// "token2" -> 127.0.0.1:8001
	// ...
	redisTokenAddressHash = "braid_token_address_hash"

	// RedisAddressTokenList 保存地址的token集合
	// 127.0.0.1:8001 -> ["token1", token2, ...]
	redisAddressTokenList = "braid_address_token_list"
)

var (
	// ErrLinkerDependRedis 使用linker需要提前初始化redis
	ErrLinkerDependRedis = errors.New("linker need depend redis")
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	defaultConfig = Config{}

	linker *Linker
)

// New 构建linker
func New() *Linker {
	linker = &Linker{}
	return linker
}

// Init 初始化linker
func (l *Linker) Init(cfg interface{}) error {
	_, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	if redis.Get() == nil {
		return ErrLinkerDependRedis
	}

	return nil
}

// Run 运行linker
func (l *Linker) Run() {

}

// Get get linker
func Get() *Linker {
	return linker
}

// Close 释放linker
func (l *Linker) Close() {

}

// Target 获取目标服务器
func (l *Linker) Target(key string) (string, error) {
	return redis.Get().HGet(redisTokenAddressHash, key)
}

// Link 将用户`标示`和包含Service的`目标服务器`进行链接，
// 在进行过链接后，下一次远程访问将会不经过coordinate，直接向目标服务器发送。
func (l *Linker) Link(key string, address string) error {
	conn := redis.Get().Conn()
	defer conn.Close()

	mu := redis.Mutex{
		Token: key,
	}
	err := mu.Lock("link")
	if err != nil {
		return err
	}
	defer mu.Unlock()

	conn.Send("MULTI")
	conn.Send("HSET", redisTokenAddressHash, key, address)
	conn.Send("LPUSH", redisAddressTokenList+address, key)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// UnLink 取消用户`标示`含有的相关链接
func (l *Linker) UnLink(token string) {

}

// Num 获取目标服务器的链接数
func (l *Linker) Num(address string) (int, error) {
	linkField := redisAddressTokenList + address
	return redis.Get().LLen(linkField)
}

// Offline 当有节点离线，将节点相关的链接进行解除
func (l *Linker) Offline(address string) error {
	client := redis.Get()
	linkField := redisAddressTokenList + address

	linkLen, err := client.LLen(linkField)
	if err != nil {
		return err
	}

	for i := 0; i < linkLen; i++ {
		key, _ := client.RPop(linkField)
		client.HDel(redisTokenAddressHash, key)
	}

	return nil
}
