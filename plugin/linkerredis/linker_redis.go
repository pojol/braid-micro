package linkerredis

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/mailbox"
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
func NewDownMsg(service string, addr string) *mailbox.Message {

	byt, _ := json.Marshal(&DownMsg{
		Service: service,
		Addr:    addr,
	})

	return &mailbox.Message{
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

func newRedisLinker() module.Builder {
	return &redisLinkerBuilder{}
}

func (*redisLinkerBuilder) Name() string {
	return Name
}

func (*redisLinkerBuilder) Type() string {
	return module.TyLinkCache
}

func (rb *redisLinkerBuilder) AddOption(opt interface{}) {
	rb.opts = append(rb.opts, opt)
}

// Build build link-cache
func (rb *redisLinkerBuilder) Build(serviceName string, mb mailbox.IMailbox) (module.IModule, error) {

	lc := &redisLinker{
		serviceName: serviceName,
		mb:          mb,
	}

	lc.mb.ClusterPub(LinkerTopicUnlink, &mailbox.Message{Body: []byte("nil")})
	lc.mb.ClusterPub(LinkerTopicDown, &mailbox.Message{Body: []byte("nil")})

	go lc.watcher()
	go lc.dispatch()

	log.Debugf("build link-cache by redis & nsq")
	return lc, nil
}

// redisLinker 基于redis实现的链接器
type redisLinker struct {
	serviceName string
	ismaster    bool
	mb          mailbox.IMailbox
	unlink      mailbox.IConsumer
	down        mailbox.IConsumer
	master      mailbox.IConsumer
}

func (l *redisLinker) Init() {

}

func (l *redisLinker) Run() {

}

func (l *redisLinker) watcher() {
	tick := time.NewTicker(time.Second)

	l.master, _ = l.mb.ProcSub(elector.BecomeMaster).AddShared()
	l.master.OnArrived(func(msg *mailbox.Message) error {
		l.ismaster = true
		l.master.Exit()
		return nil
	})

	for {
		select {
		case <-tick.C:
			if l.ismaster {

				l.unlink, _ = l.mb.ClusterSub(LinkerTopicUnlink).AddShared()
				l.unlink.OnArrived(func(msg *mailbox.Message) error {
					return l.Unlink(string(msg.Body))
				})

				l.down, _ = l.mb.ClusterSub(LinkerTopicDown).AddShared()
				l.down.OnArrived(func(msg *mailbox.Message) error {

					downMsg := &DownMsg{}
					err := json.Unmarshal(msg.Body, downMsg)
					if err != nil {
						return nil
					}
					return l.Down(discover.Node{
						Name:    downMsg.Service,
						Address: downMsg.Addr,
					})
				})

				tick.Stop()
				return
			}
		}
	}

}

func (l *redisLinker) dispatchLinkinfo() {
	keys, _ := redis.Get().Keys(LinkerRedisPrefix + "lst-*")
	for _, key := range keys {
		nod := strings.Split(key, "-")
		num, _ := redis.Get().LLen(key)

		parent := nod[2]
		//child := nod[3]
		id := nod[4]

		if l.serviceName == parent {
			l.mb.ProcPub(linkcache.ServiceLinkNum, linkcache.EncodeLinkNumMsg(id, num))
		}

	}
}

func (l *redisLinker) dispatch() {
	tick := time.NewTicker(time.Second * 3)

	for {
		select {
		case <-tick.C:
			l.dispatchLinkinfo()
		}
	}
}

func (l *redisLinker) Target(token string, serviceName string) (target string, err error) {

	val, err := redis.Get().HGet(LinkerRedisPrefix+"hash", l.serviceName+"-"+token)
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

	val, err := redis.ConnHGet(conn, LinkerRedisPrefix+"hash", l.serviceName+"-"+token)
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
	conn.Send("SADD", LinkerRedisPrefix+"relation-"+l.serviceName, cia)
	conn.Send("LPUSH", LinkerRedisPrefix+"lst-"+l.serviceName+"-"+cia, token)
	conn.Send("HSET", LinkerRedisPrefix+"hash", l.serviceName+"-"+token, string(dat))

	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	log.Debugf("linked parent %s, target %s, token %s", l.serviceName, target.Address, token)
	return nil
}

// Unlink 当前节点所属的用户离线
func (l *redisLinker) Unlink(token string) error {

	if token == "" {
		return nil
	}

	conn := redis.Get().Conn()
	defer conn.Close()

	val, err := redis.ConnHGet(conn, LinkerRedisPrefix+"hash", l.serviceName+"-"+token)
	if err != nil {
		return err
	}
	if val == "" {
		return nil
	}

	tr := tokenRelation{}
	err = json.Unmarshal([]byte(val), &tr.targets)
	if err != nil {
		fmt.Println(string(val), err, "field", LinkerRedisPrefix+"hash", l.serviceName+"-"+token)
		return err
	}

	conn.Send("MULTI")
	for _, t := range tr.targets {
		cia := t.Name + "-" + t.ID + "-" + t.Address
		conn.Send("LREM", LinkerRedisPrefix+"lst-"+l.serviceName+"-"+cia, 0, token)
	}
	conn.Send("HDEL", LinkerRedisPrefix+"hash", l.serviceName+"-"+token)

	_, err = conn.Do("EXEC")
	if err != nil {
		log.Debugf("unlink exec err %s", err.Error())
		return err
	}

	return nil
}

// Down 删除离线节点的链路缓存
func (l *redisLinker) Down(target discover.Node) error {

	conn := redis.Get().Conn()
	defer conn.Close()

	cia := target.Name + "-" + target.ID + "-" + target.Address
	ismember, err := redis.ConnSIsMember(conn, LinkerRedisPrefix+"relation-"+l.serviceName, cia)
	if err != nil {
		return err
	}
	if !ismember {
		return nil
	}

	log.Debugf("redis linker down child %s, target %s", target.Name, target.Address)

	tokens, err := redis.ConnLRange(conn, LinkerRedisPrefix+"lst-"+l.serviceName+"-"+cia, 0, -1)
	if err != nil {
		log.Debugf("linker down ConnLRange %s", err.Error())
		return err
	}

	conn.Send("MULTI")
	for _, token := range tokens {
		conn.Send("HDEL", LinkerRedisPrefix+"hash", l.serviceName+"-"+token)
	}
	conn.Send("DEL", LinkerRedisPrefix+"lst-"+l.serviceName+"-"+cia)
	conn.Send("SREM", LinkerRedisPrefix+"relation-"+l.serviceName, cia)
	_, err = conn.Do("EXEC")
	if err != nil {
		log.Debugf("linker down exec %s", err.Error())
		return err
	}

	return nil
}

func (l *redisLinker) Close() {

}

func init() {
	module.Register(newRedisLinker())
}
