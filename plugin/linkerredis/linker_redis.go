package linkerredis

import (
	"errors"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/pubsub"
)

var (
	// LinkerRedisPrefix linker redis key prefix
	LinkerRedisPrefix = "braid_linker_"
)

const (
	// LinkerName 链接器名称
	LinkerName = "RedisLinker"

	// redis

	// braid_linker_"parent"_"chlid"_"addr" `list` 本节点链路下的目标节点的token集合
	// parent_child_addr => [token ...]

	// braid_linker_"parent"_"token"  `set` token在本节点下面进行链路的目标节点集合
	// parent_token => ( targetNod ... )

	// LinkerRedisTokenPool token pool
	// braid_linker_"parent"_"child"_"token" `hash` 本节点下token指向的目标地址
	// parent_chlid_token : "targetAddr"
	LinkerRedisTokenPool = "token_pool"

	// topic

	// LinkerTopicUnlink unlink token topic
	LinkerTopicUnlink = "braid_linker_unlink"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

type redisLinkerBuilder struct {
	cfg Config
}

func newRedisLinker() linkcache.Builder {
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

func (rb *redisLinkerBuilder) Build(elector elector.IElection, ps pubsub.IPubsub) linkcache.ILinkCache {

	e := &redisLinker{
		elector: elector,
		pubsub:  ps,
		cfg:     rb.cfg,
	}

	unlinkSub := ps.Sub(LinkerTopicUnlink)
	unlinkSub.AddShared().OnArrived(func(msg *pubsub.Message) error {

		if elector.IsMaster() {
			e.Unlink(string(msg.Body))
		}

		return nil
	})

	log.Debugf("build redis linker")
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

func (l *redisLinker) getTokenPoolField(child string, token string) string {
	field := LinkerRedisPrefix + l.cfg.ServiceName + "_" + child + "_" + token
	return field
}

func getTokenPoolKey() string {
	return LinkerRedisPrefix + LinkerRedisTokenPool
}

func (l *redisLinker) getTokenListField(child string, addr string) string {
	field := LinkerRedisPrefix + "lst_" + l.cfg.ServiceName + "_" + child + "_" + addr
	return field
}

func (l *redisLinker) getParentSetField(token string) string {
	field := LinkerRedisPrefix + "set_" + l.cfg.ServiceName + "_" + token
	return field
}

func (l *redisLinker) Target(child string, token string) (target string, err error) {

	if token == "" {
		return "", nil
	}

	return redis.Get().HGet(getTokenPoolKey(), l.getTokenPoolField(child, token))
}

func (l *redisLinker) Link(child string, token string, targetAddr string) error {

	conn := redis.Get().Conn()
	defer conn.Close()

	mu := redis.Mutex{
		Token: token,
	}
	err := mu.Lock("braid_linker")
	if err != nil {
		return err
	}
	defer mu.Unlock()

	conn.Send("MULTI")
	conn.Send("HSET", getTokenPoolKey(), l.getTokenPoolField(child, token), targetAddr)
	conn.Send("SADD", l.getParentSetField(token), child)
	conn.Send("LPUSH", l.getTokenListField(child, targetAddr), token)
	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	log.Debugf("linked parent %s, target %s, token %s", l.cfg.ServiceName, targetAddr, token)
	return nil
}

// Unlink 当前节点所属的用户离线
func (l *redisLinker) Unlink(token string) error {

	if token == "" {
		return nil
	}

	conn := redis.Get().Conn()
	defer conn.Close()

	childs, err := redis.ConnSMembers(conn, l.getParentSetField(token))
	if err != nil {
		log.Debugf("unlink get parent members err %s", err.Error())
		return err
	}

	targets := []string{}

	for _, child := range childs {

		targetAddr, err := redis.ConnHGet(conn, getTokenPoolKey(), l.getTokenPoolField(child, token))
		if err != nil {
			log.Debugf("unlink hget %s", err.Error())
			continue
		}

		targets = append(targets, targetAddr)
	}

	conn.Send("MULTI")

	for i := 0; i < len(childs); i++ {
		child := childs[i]
		targetAddr := targets[i]

		conn.Send("HDEL", getTokenPoolKey(), l.getTokenPoolField(child, token))
		if targetAddr != "" {
			conn.Send("LREM", l.getTokenListField(child, targetAddr), 0, token)
		}

		log.Debugf("unlinked parent %s, target %s, token %s", l.cfg.ServiceName, targetAddr, token)
	}

	conn.Send("DEL", l.getParentSetField(token))

	_, err = conn.Do("EXEC")
	if err != nil {
		log.Debugf("unlink exec err %s", err.Error())
	}

	return nil
}

func (l *redisLinker) Num(child string, targetAddr string) (int, error) {
	return redis.Get().LLen(l.getTokenListField(child, targetAddr))
}

// Down 删除离线节点的链路缓存
func (l *redisLinker) Down(child string, targetAddr string) error {

	if !l.elector.IsMaster() {
		return nil
	}

	conn := redis.Get().Conn()
	defer conn.Close()

	tokens, err := redis.ConnLRange(conn, l.getTokenListField(child, targetAddr), 0, -1)
	if err != nil {
		log.Debugf("linker down ConnLRange %s", err.Error())
		return err
	}

	conn.Send("MULTI")
	for _, token := range tokens {
		conn.Send("HDEL", getTokenPoolKey(), l.getTokenPoolField(child, token))
	}
	conn.Send("DEL", l.getTokenListField(child, targetAddr))
	_, err = conn.Do("EXEC")
	if err != nil {
		log.Debugf("linker down exec %s", err.Error())
		return err
	}

	return nil
}

func init() {
	linkcache.Register(newRedisLinker())
}
