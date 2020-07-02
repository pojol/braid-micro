package redislinker

import (
	"context"
	"errors"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/module/tracer"
	"github.com/pojol/braid/plugin/linker"
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
	cecfg, ok := cfg.(Cfg)
	if !ok {
		return nil
	}

	e := &redisLinker{
		cfg: cecfg,
	}

	return e
}

// Cfg linker config
type Cfg struct {
	Tracing bool
}

// redisLinker 基于redis实现的链接器
type redisLinker struct {
	cfg Cfg
}

func (l *redisLinker) Target(ctx context.Context, token string) (target string, err error) {

	if l.cfg.Tracing {
		tr := tracer.RedisTracer{
			Cmd: "HGET",
		}
		tr.Begin(ctx)
		defer tr.End()
	}

	return redis.Get().HGet(TokenAddressHash, token)
}

func (l *redisLinker) Link(ctx context.Context, token string, nodid string, target string) error {

	conn := redis.Get().Conn()
	defer conn.Close()

	if l.cfg.Tracing {
		rt := tracer.RedisTracer{
			Cmd: "HSET|LPUSH",
		}
		rt.Begin(ctx)
		defer rt.End()
	}

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

func (l *redisLinker) Unlink(token string) error {

	return nil
}

func (l *redisLinker) Num(ctx context.Context, nodid string) (int, error) {
	linkField := TokenList + "_" + nodid

	if l.cfg.Tracing {
		rt := tracer.RedisTracer{
			Cmd: "LLEN",
		}
		rt.Begin(ctx)
		defer rt.End()
	}

	return redis.Get().LLen(linkField)
}

func (l *redisLinker) Offline(ctx context.Context, nodid string) error {
	client := redis.Get()
	linkField := TokenList + "_" + nodid

	if l.cfg.Tracing {
		rt := tracer.RedisTracer{
			Cmd: "LLEN | RPOP",
		}
		rt.Begin(ctx)
		defer rt.End()
	}

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
