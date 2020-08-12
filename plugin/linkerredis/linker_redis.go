package linkerredis

import (
	"errors"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/module/linker"
)

const (
	// LinkerName 链接器名称
	LinkerName = "RedisLinker"

	// TokenAddressHash hash { key : token , val : target }
	TokenAddressHash = "braid_token_address_hash"

	// TokenList list ["token" ...]
	TokenList = "braid_token_list"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

type redisLinkerBuilder struct{}

func newRedisLinker() linker.Builder {
	return &redisLinkerBuilder{}
}

func (*redisLinkerBuilder) Name() string {
	return LinkerName
}

func (*redisLinkerBuilder) Build(cfg interface{}) linker.ILinker {

	e := &redisLinker{}

	return e
}

// redisLinker 基于redis实现的链接器
type redisLinker struct {
}

func (l *redisLinker) Target(token string) (target string, err error) {

	if token == "" {
		return "", nil
	}

	return redis.Get().HGet(TokenAddressHash, token)
}

func (l *redisLinker) Link(token string, nodid string, target string) error {

	conn := redis.Get().Conn()
	defer conn.Close()

	mu := redis.Mutex{
		Token: token,
	}
	err := mu.Lock("Link")
	if err != nil {
		return err
	}
	defer mu.Unlock()

	conn.Send("MULTI")
	conn.Send("HSET", TokenAddressHash, token, target)
	conn.Send("LPUSH", TokenList+"_"+nodid, token)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// Unlink 当前节点所属的用户离线
func (l *redisLinker) Unlink(token string) error {

	if token == "" {
		return nil
	}

	return nil
}

func (l *redisLinker) Num(nodid string) (int, error) {
	linkField := TokenList + "_" + nodid
	return redis.Get().LLen(linkField)
}

// Offline 删除离线节点的链路缓存
// 注: 这个函数调用可能会被多个节点调用
func (l *redisLinker) Offline(nodid string) error {
	client := redis.Get()
	linkField := TokenList + "_" + nodid

	linkLen, err := client.LLen(linkField)
	if err != nil {
		return err
	}

	for i := 0; i < linkLen; i++ {
		token, _ := client.RPop(linkField)
		client.HDel(TokenAddressHash, token)
	}

	return nil
}

func init() {
	linker.Register(newRedisLinker())
}
