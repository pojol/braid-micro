package linkerredis

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/pubsub"
)

var (
	// LinkerRedisPrefix linker redis key prefix
	LinkerRedisPrefix = "braid_linker-"
)

const (
	// Name 链接器名称
	Name = "RedisLinker"

	// topic

	// LinkerTopicUnlink unlink token topic
	LinkerTopicUnlink = "braid_linker_unlink"
	// LinkerTopicDown service down
	LinkerTopicDown = "braid_linker_down"
)

// DownMsg down msg
type DownMsg struct {
	Service string
	Addr    string
}

type tokenRelation struct {
	targets []discover.Node
}

// NewDownMsg new down msg
func NewDownMsg(service string, addr string) *pubsub.Message {

	byt, _ := json.Marshal(&DownMsg{
		Service: service,
		Addr:    addr,
	})

	return &pubsub.Message{
		Body: byt,
	}
}

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

type redisLinkerBuilder struct {
	opts []interface{}
}

func newRedisLinker() linkcache.Builder {
	return &redisLinkerBuilder{}
}

func (*redisLinkerBuilder) Name() string {
	return Name
}

func (rb *redisLinkerBuilder) AddOption(opt interface{}) {
	rb.opts = append(rb.opts, opt)
}

// Build build link-cache
func (rb *redisLinkerBuilder) Build(serviceName string) (linkcache.ILinkCache, error) {

	p := Parm{
		ServiceName: serviceName,
	}
	for _, opt := range rb.opts {
		opt.(Option)(&p)
	}

	if p.elector == nil {
		return nil, errors.New("linker_redis parm mismatch " + "no elector")
	}
	if p.clusterPB == nil {
		return nil, errors.New("linker_redis parm mismatch " + "no cluster pubsub")
	}

	lc := &redisLinker{
		parm: p,
	}

	lc.parm.clusterPB.Pub(LinkerTopicUnlink, &pubsub.Message{Body: []byte("nil")})
	lc.parm.clusterPB.Pub(LinkerTopicDown, &pubsub.Message{Body: []byte("nil")})

	go lc.watcher()

	log.Debugf("build link-cache by redis & nsq")
	return lc, nil
}

// redisLinker 基于redis实现的链接器
type redisLinker struct {
	parm   Parm
	unlink pubsub.IConsumer
	down   pubsub.IConsumer
}

func (l *redisLinker) watcher() {
	tick := time.NewTicker(time.Second)

	for {
		select {
		case <-tick.C:
			if l.parm.elector.IsMaster() {

				l.unlink = l.parm.clusterPB.Sub(LinkerTopicUnlink).AddShared()
				l.unlink.OnArrived(func(msg *pubsub.Message) error {

					if l.parm.elector.IsMaster() {
						return l.Unlink(string(msg.Body))
					}

					return nil
				})

				l.down = l.parm.clusterPB.Sub(LinkerTopicDown).AddShared()
				l.down.OnArrived(func(msg *pubsub.Message) error {

					if l.parm.elector.IsMaster() {
						downMsg := &DownMsg{}
						err := json.Unmarshal(msg.Body, downMsg)
						if err != nil {
							return nil
						}
						return l.Down(discover.Node{
							Name:    downMsg.Service,
							Address: downMsg.Addr,
						})
					}
					return nil
				})

				tick.Stop()
				return
			}
		}
	}

}

func (l *redisLinker) Target(token string, serviceName string) (target string, err error) {

	val, err := redis.Get().HGet(LinkerRedisPrefix+"hash", l.parm.ServiceName+"-"+token)
	if err != nil {
		return "", err
	}

	if val == "" {
		return "", nil
	}

	tr := tokenRelation{}
	err = json.Unmarshal([]byte(val), &tr.targets)
	if err != nil {
		return "", err
	}

	for _, v := range tr.targets {
		if v.Name == serviceName {
			return v.Address, nil
		}
	}

	return "", nil
}

func (l *redisLinker) Link(token string, target discover.Node) error {

	conn := redis.Get().Conn()
	defer conn.Close()

	mu := redis.Mutex{
		Token: token,
	}
	err := mu.Lock("braid_link_token")
	if err != nil {
		return err
	}
	defer mu.Unlock()

	val, err := redis.ConnHGet(conn, LinkerRedisPrefix+"hash", l.parm.ServiceName+"-"+token)
	if err != nil {
		return err
	}

	tr := tokenRelation{}
	var dat []byte
	if val != "" {
		err = json.Unmarshal([]byte(val), &tr.targets)
		if err != nil {
			return err
		}
	}

	tr.targets = append(tr.targets, target)
	dat, err = json.Marshal(&tr.targets)
	if err != nil {
		return err
	}

	cia := target.Name + "-" + target.ID + "-" + target.Address
	conn.Send("MULTI")
	conn.Send("SADD", LinkerRedisPrefix+"relation-"+l.parm.ServiceName, cia)
	conn.Send("LPUSH", LinkerRedisPrefix+"lst-"+l.parm.ServiceName+"-"+cia, token)
	conn.Send("HSET", LinkerRedisPrefix+"hash", l.parm.ServiceName+"-"+token, string(dat))

	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	log.Debugf("linked parent %s, target %s, token %s", l.parm.ServiceName, target.Address, token)
	return nil
}

// Unlink 当前节点所属的用户离线
func (l *redisLinker) Unlink(token string) error {

	if token == "" {
		return nil
	}

	conn := redis.Get().Conn()
	defer conn.Close()

	val, err := redis.ConnHGet(conn, LinkerRedisPrefix+"hash", l.parm.ServiceName+"-"+token)
	if err != nil {
		return err
	}
	if val == "" {
		return nil
	}

	tr := tokenRelation{}
	err = json.Unmarshal([]byte(val), &tr.targets)
	if err != nil {
		fmt.Println(string(val), err, "field", LinkerRedisPrefix+"hash", l.parm.ServiceName+"-"+token)
		return err
	}

	conn.Send("MULTI")
	for _, t := range tr.targets {
		cia := t.Name + "-" + t.ID + "-" + t.Address
		conn.Send("LREM", LinkerRedisPrefix+"lst-"+l.parm.ServiceName+"-"+cia, 0, token)
	}
	conn.Send("HDEL", LinkerRedisPrefix+"hash", l.parm.ServiceName+"-"+token)

	_, err = conn.Do("EXEC")
	if err != nil {
		log.Debugf("unlink exec err %s", err.Error())
		return err
	}

	return nil
}

func (l *redisLinker) Num(target discover.Node) (int, error) {
	field := LinkerRedisPrefix + "lst-" + l.parm.ServiceName + "-" + target.Name + "-" + target.ID + "_" + target.Address
	return redis.Get().LLen(field)
}

// Down 删除离线节点的链路缓存
func (l *redisLinker) Down(target discover.Node) error {

	conn := redis.Get().Conn()
	defer conn.Close()

	cia := target.Name + "-" + target.ID + "-" + target.Address
	ismember, err := redis.ConnSIsMember(conn, LinkerRedisPrefix+"relation-"+l.parm.ServiceName, cia)
	if err != nil {
		return err
	}
	if !ismember {
		return nil
	}

	log.Debugf("redis linker down child %s, target %s", target.Name, target.Address)

	tokens, err := redis.ConnLRange(conn, LinkerRedisPrefix+"lst-"+l.parm.ServiceName+"-"+cia, 0, -1)
	if err != nil {
		log.Debugf("linker down ConnLRange %s", err.Error())
		return err
	}

	conn.Send("MULTI")
	for _, token := range tokens {
		conn.Send("HDEL", LinkerRedisPrefix+"hash", l.parm.ServiceName+"-"+token)
	}
	conn.Send("DEL", LinkerRedisPrefix+"lst-"+l.parm.ServiceName+"-"+cia)
	conn.Send("SREM", LinkerRedisPrefix+"relation-"+l.parm.ServiceName, cia)
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
