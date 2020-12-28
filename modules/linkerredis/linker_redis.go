package linkerredis

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/logger"
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

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

type (

	// RedisClient redis client
	RedisClient struct {
		pool    *redis.Pool
		Address string
	}

	// RedisConfig 配置项
	RedisConfig struct {
		Address string //connection string, like "redis:// :password@10.0.1.11:6379/0"

		ReadTimeOut    time.Duration // 连接的读取超时时间
		WriteTimeOut   time.Duration // 连接的写入超时时间
		ConnectTimeOut time.Duration // 连接超时时间
		MaxIdle        int           // 最大空闲连接数
		MaxActive      int           // 最大连接数，当为0时没有连接数限制
		IdleTimeout    time.Duration // 闲置连接的超时时间, 设置小于服务器的超时时间 redis.conf : timeout
	}
)

// Ping 测试一个连接是否可用
func (rc *RedisClient) Ping() (string, error) {
	conn := rc.pool.Get()
	defer conn.Close()
	val, err := redis.String(conn.Do("PING"))
	return val, err
}

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
func (rb *redisLinkerBuilder) Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (module.IModule, error) {

	p := Parm{
		RedisAddr:      "redis://127.0.0.1:6379/0",
		RedisMaxIdle:   16,
		RedisMaxActive: 128,
	}
	for _, opt := range rb.opts {
		opt.(Option)(&p)
	}

	client := &RedisClient{
		pool: &redis.Pool{
			MaxIdle:   p.RedisMaxIdle,
			MaxActive: p.RedisMaxActive,
			Dial: func() (redis.Conn, error) {
				c, err := redis.DialURL(
					p.RedisAddr,
					redis.DialReadTimeout(5*time.Second),
					redis.DialWriteTimeout(5*time.Second),
					redis.DialConnectTimeout(2*time.Second),
				)
				return c, err
			},
		},
	}

	_, err := client.Ping()
	if err != nil {
		logger.Debugf("redis ping err %s", err.Error())
		return nil, err
	}

	lc := &redisLinker{
		serviceName:  serviceName,
		mb:           mb,
		electorState: elector.EWait,
		logger:       logger,
		client:       client,
	}

	lc.mb.PubAsync(mailbox.Cluster, LinkerTopicUnlink, &mailbox.Message{Body: []byte("nil")})
	lc.mb.PubAsync(mailbox.Cluster, LinkerTopicDown, linkcache.EncodeDownMsg("", "", ""))

	go lc.watcher()
	go lc.dispatch()

	logger.Debugf("build link-cache by redis & nsq")
	return lc, nil
}

// redisLinker 基于redis实现的链接器
type redisLinker struct {
	serviceName  string
	electorState string
	mb           mailbox.IMailbox
	unlink       mailbox.IConsumer
	down         mailbox.IConsumer

	logger logger.ILogger

	client *RedisClient

	sync.Mutex
}

func (l *redisLinker) Init() {

}

func (l *redisLinker) Run() {

}

func (l *redisLinker) watcher() {

	l.unlink, _ = l.mb.Sub(mailbox.Cluster, LinkerTopicUnlink).Competition()
	l.down, _ = l.mb.Sub(mailbox.Cluster, LinkerTopicDown).Competition()

	l.unlink.OnArrived(func(msg mailbox.Message) error {
		l.logger.Debugf("recv unlink msg %s", string(msg.Body))
		return l.Unlink(string(msg.Body), "")
	})

	l.down.OnArrived(func(msg mailbox.Message) error {
		dmsg := linkcache.DecodeDownMsg(&msg)
		if dmsg.Service == "" {
			return errors.New("Can't find service")
		}

		return l.Down(discover.Node{
			ID:      dmsg.ID,
			Name:    dmsg.Service,
			Address: dmsg.Addr,
		})
	})

}

func (l *redisLinker) getConn() redis.Conn {
	return l.client.pool.Get()
}

func (l *redisLinker) dispatchLinkinfo() {

	conn := l.getConn()
	defer conn.Close()

	keys, _ := redis.Strings(conn.Do("KEYS", LinkerRedisPrefix+"lst-*"))
	for _, key := range keys {
		nod := strings.Split(key, "-")
		num, _ := redis.Int(conn.Do("LLen", key))

		parent := nod[2]
		//child := nod[3]
		id := nod[4]

		if l.serviceName == parent {
			l.mb.Pub(mailbox.Proc, linkcache.ServiceLinkNum, linkcache.EncodeLinkNumMsg(id, num))
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

	conn := l.getConn()
	defer conn.Close()

	val, err := redis.String(conn.Do("HGET", LinkerRedisPrefix+"hash", l.serviceName+"-"+token+"-"+serviceName))
	if val == "" || err != nil {
		return "", nil
	}

	return val, nil
}

func (l *redisLinker) Link(token string, target discover.Node) error {
	l.Lock()
	defer l.Unlock()

	conn := l.getConn()
	defer conn.Close()

	cia := target.Name + "-" + target.ID
	conn.Send("MULTI")
	conn.Send("SADD", LinkerRedisPrefix+"relation-"+l.serviceName+"-"+token, cia)
	conn.Send("SADD", LinkerRedisPrefix+"relation-"+l.serviceName, cia)
	conn.Send("LPUSH", LinkerRedisPrefix+"lst-"+l.serviceName+"-"+cia, token)
	conn.Send("HSET", LinkerRedisPrefix+"hash", l.serviceName+"-"+token+"-"+target.Name, target.Address)

	_, err := conn.Do("EXEC")
	if err != nil {
		l.logger.Debugf("link exec err %s", err.Error())
		return err
	}

	l.logger.Debugf("linked parent %s, target %s, token %s", l.serviceName, cia, token)
	return nil
}

// Unlink 当前节点所属的用户离线
func (l *redisLinker) Unlink(token string, target string) error {
	if token == "" || token == "nil" {
		return nil
	}

	l.Lock()
	defer l.Unlock()

	conn := l.getConn()
	defer conn.Close()
	relations, err := redis.Strings(conn.Do("SMEMBERS", LinkerRedisPrefix+"relation-"+l.serviceName+"-"+token))
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	lremCount := 0

	if target != "" {

		for _, relation := range relations {
			rinfo := strings.Split(relation, "-")
			if rinfo[0] == target {
				conn.Send("LREM", LinkerRedisPrefix+"lst-"+l.serviceName+"-"+relation, 0, token)
				lremCount++
			}
		}

		conn.Send("HDEL", LinkerRedisPrefix+"hash", l.serviceName+"-"+token+"-"+target)

		if lremCount == len(relations) {
			conn.Send("DEL", LinkerRedisPrefix+"relation-"+l.serviceName+"-"+token)
		}

	} else {

		for _, relation := range relations {
			rinfo := strings.Split(relation, "-")
			conn.Send("LREM", LinkerRedisPrefix+"lst-"+l.serviceName+"-"+relation, 0, token)
			conn.Send("HDEL", LinkerRedisPrefix+"hash", l.serviceName+"-"+token+"-"+rinfo[0])
		}

		conn.Send("DEL", LinkerRedisPrefix+"relation-"+l.serviceName+"-"+token)
	}

	_, err = conn.Do("EXEC")
	if err != nil {
		l.logger.Debugf("unlink exec err %s", err.Error())
		return err
	}

	l.logger.Debugf("unlink token service: %s, target :%s, token :%s", l.serviceName, target, token)
	return nil
}

// Down 删除离线节点的链路缓存
func (l *redisLinker) Down(target discover.Node) error {

	conn := l.getConn()
	defer conn.Close()
	cia := target.Name + "-" + target.Address

	ismember, err := redis.Bool(conn.Do("SISMEMBER", LinkerRedisPrefix+"relation-"+l.serviceName, cia))
	if err != nil {
		return err
	}
	if !ismember {
		return nil
	}

	l.logger.Debugf("redis linker down child %s, target %s", target.Name, target.Address)

	tokens, err := redis.Strings(conn.Do("LRANGE", LinkerRedisPrefix+"lst-"+l.serviceName+"-"+cia, 0, -1))
	if err != nil {
		l.logger.Debugf("linker down ConnLRange %s", err.Error())
		return err
	}

	conn.Send("MULTI")
	for _, token := range tokens {
		conn.Do("HDEL", LinkerRedisPrefix+"hash", l.serviceName+"-"+token+"-"+target.Name)
		conn.Do("SREM", LinkerRedisPrefix+"relation-"+l.serviceName+"-"+token, target.Name+"-"+target.ID)
	}
	conn.Send("DEL", LinkerRedisPrefix+"lst-"+l.serviceName+"-"+cia)

	_, err = conn.Do("EXEC")
	if err != nil {
		l.logger.Debugf("linker down exec %s", err.Error())
		return err
	}

	return nil
}

func (l *redisLinker) Close() {
	l.client.pool.Close()
}

func createTopic(topic string) {

}

func init() {
	module.Register(newRedisLinker())
}
