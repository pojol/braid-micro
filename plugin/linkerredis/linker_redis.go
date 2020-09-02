package linkerredis

import (
	"errors"
	"fmt"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linker"
	"github.com/pojol/braid/module/pubsub"
)

const (
	// LinkerName 链接器名称
	LinkerName = "RedisLinker"

	// LinkerPrefix linker redis key prefix
	LinkerPrefix = "braid_linker_"

	// braid_linker_"base"_child_"addr" [ token ... ]

	// LinkerTokenPool token pool
	LinkerTokenPool = "braid_linker_token_pool"
	// braid_linker_"base"_"token" : "targetAddr"

	// LinkerDownTopic node offline topic
	LinkerDownTopic = "braid_linker_service_down"

	// LinkerUnlinkTopic unlink token topic
	LinkerUnlinkTopic = "braid_linker_unlink"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

type redisLinkerBuilder struct {
	cfg Config
}

func newRedisLinker() linker.Builder {
	return &redisLinkerBuilder{}
}

func (*redisLinkerBuilder) Name() string {
	return LinkerName
}

func (rb *redisLinkerBuilder) SetCfg(cfg interface{}) error {
	lcfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	rb.cfg = lcfg
	return nil
}

func (rb *redisLinkerBuilder) Build(elector elector.IElection, ps pubsub.IPubsub) linker.ILinker {

	e := &redisLinker{
		elector: elector,
		pubsub:  ps,
		cfg:     rb.cfg,
	}

	downSub := ps.Sub(LinkerDownTopic)
	downSub.AddShared().OnArrived(func(msg *pubsub.Message) error {

		// 本节点是master节点，才可以执行下面的操作.
		if elector.IsMaster() {
			e.Down(string(msg.Body))
		}

		return nil
	})

	unlinkSub := ps.Sub(LinkerUnlinkTopic)
	unlinkSub.AddShared().OnArrived(func(msg *pubsub.Message) error {

		if elector.IsMaster() {
			e.Unlink(string(msg.Body))
		}

		return nil
	})

	return e
}

// Config linker config
type Config struct {
	ServiceName string
}

// redisLinker 基于redis实现的链接器
type redisLinker struct {
	elector elector.IElection
	pubsub  pubsub.IPubsub
	cfg     Config
}

func (l *redisLinker) Target(token string) (target string, err error) {

	if token == "" {
		return "", nil
	}

	childField := LinkerPrefix + l.cfg.ServiceName + "_" + token

	return redis.Get().HGet(LinkerTokenPool, childField)
}

func (l *redisLinker) Link(token string, targetAddr string) error {

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

	parent := l.cfg.ServiceName

	conn.Send("MULTI")
	conn.Send("HSET", LinkerTokenPool, LinkerPrefix+parent+"_"+token, targetAddr)
	conn.Send("LPUSH", LinkerPrefix+parent+"_child_"+targetAddr, token)
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

	conn := redis.Get().Conn()
	defer conn.Close()

	childField := LinkerPrefix + l.cfg.ServiceName + "_" + token

	targetAddr, err := redis.ConnHGet(conn, LinkerTokenPool, childField)
	if err != nil {
		fmt.Println("unlink hget", err)
	}

	addrField := LinkerPrefix + l.cfg.ServiceName + "_child_" + targetAddr

	conn.Send("MULTI")
	conn.Send("HDEL", LinkerTokenPool, childField)
	conn.Send("LREM", addrField, 0, token)
	_, err = conn.Do("EXEC")
	if err != nil {
		fmt.Println("unlink exec", err)
	}

	return nil
}

func (l *redisLinker) Num(targetAddr string) (int, error) {
	linkField := LinkerPrefix + l.cfg.ServiceName + "_child_" + targetAddr
	return redis.Get().LLen(linkField)
}

// Down 删除离线节点的链路缓存
func (l *redisLinker) Down(targetAddr string) error {

	return nil
}

func init() {
	linker.Register(newRedisLinker())
}
