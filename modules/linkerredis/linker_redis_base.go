package linkerredis

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
)

var (
	// LinkerRedisPrefix linker redis key prefix
	LinkerRedisPrefix = "braid_linker-"

	// RelationPrefix braid_linker-relation-parent-child : cnt
	RelationPrefix = LinkerRedisPrefix + relationFlag

	// RoutePrefix braid_linker-route-gate-base : nodinfo { addr, name, id }
	RoutePrefix = LinkerRedisPrefix + routeFlag
)

const (
	// Name 链接器名称
	Name = "RedisLinker"

	splitFlag = "-"

	// sankey
	//braid_linker-relation-parent-child : cnt
	relationFlag = "relation"

	// braid_linker-route-gate-base : nodinfo { addr, name, id }
	// 这个字段用于描述 父-子 节点之间的链路关系，通常用在随机请求端
	routeFlag = "route"

	// braid_linker-linknum-gate-base-ID : 100
	linknumFlag = "linknum"
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
		Mode:           LinkerRedisModeRedis,
		SyncTick:       1000 * 10, // 10 second
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

	lc := &redisLinker{
		serviceName:  serviceName,
		mb:           mb,
		electorState: elector.EWait,
		logger:       logger,
		client:       client,
		parm:         p,
		local: &localLinker{
			serviceName: serviceName,
			tokenMap:    make(map[string]linkInfo),
			relationSet: make(map[string]int),
		},
	}

	lc.mb.PubAsync(mailbox.Cluster, linkcache.TopicUnlink, linkcache.EncodeUnlinkMsg("", ""))
	lc.mb.PubAsync(mailbox.Cluster, linkcache.TopicDown, linkcache.EncodeDownMsg("", "", ""))

	return lc, nil
}

type linkInfo struct {
	TargetAddr string
	TargetID   string
	TargetName string
}

// redisLinker 基于redis实现的链接器
type redisLinker struct {
	serviceName string
	parm        Parm

	electorState string
	mb           mailbox.IMailbox

	unlink   mailbox.IConsumer
	down     mailbox.IConsumer
	election mailbox.IConsumer

	logger logger.ILogger

	client *RedisClient
	local  *localLinker

	sync.RWMutex
}

func (rl *redisLinker) Init() error {
	var err error
	rl.unlink, err = rl.mb.Sub(mailbox.Cluster, linkcache.TopicUnlink).Shared()
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", rl.serviceName, "mailbox", linkcache.TopicUnlink)
	}

	rl.down, err = rl.mb.Sub(mailbox.Cluster, linkcache.TopicDown).Shared()
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", rl.serviceName, "mailbox", linkcache.TopicDown)
	}

	rl.election, err = rl.mb.Sub(mailbox.Proc, elector.StateChange).Shared()
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", rl.serviceName, "elector", elector.StateChange)
	}

	_, err = rl.client.Ping()
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", rl.serviceName, "redis", rl.parm.RedisAddr)
	}

	rl.unlink.OnArrived(func(msg mailbox.Message) error {

		unlinkmsg := linkcache.DecodeUnlinkMsg(&msg)
		if unlinkmsg.Token != "" {
			return rl.Unlink(unlinkmsg.Token, unlinkmsg.Service)
		}

		return nil
	})

	rl.down.OnArrived(func(msg mailbox.Message) error {
		dmsg := linkcache.DecodeDownMsg(&msg)
		if dmsg.Service == "" {
			return errors.New("Can't find service")
		}

		return rl.Down(discover.Node{
			ID:      dmsg.ID,
			Name:    dmsg.Service,
			Address: dmsg.Addr,
		})
	})

	rl.election.OnArrived(func(msg mailbox.Message) error {

		statemsg := elector.DecodeStateChangeMsg(&msg)
		if statemsg.State != "" {
			rl.electorState = statemsg.State
		}

		return nil
	})

	return nil
}

func (rl *redisLinker) syncLinkNum() {
	conn := rl.getConn()
	defer conn.Close()

	members, err := redis.Strings(conn.Do("SMEMBERS", RelationPrefix))
	if err != nil {
		return
	}

	for _, member := range members {
		info := strings.Split(member, splitFlag)
		if len(info) != 5 {
			rl.logger.Warnf("%v wrong relation string format %v", Name, member)
			continue
		}

		parent := info[2]
		id := info[4]
		if rl.serviceName != parent {
			continue
		}

		cnt, err := redis.Int64(conn.Do("GET", member))
		if err != nil {
			rl.logger.Warnf("%v redis cmd err %v", Name, err.Error())
			continue
		}

		rl.mb.Pub(mailbox.Cluster, linkcache.ServiceLinkNum, linkcache.EncodeLinkNumMsg(id, int(cnt)))
	}
}

func (rl *redisLinker) Run() {

	// 这里还要处理下 历史数据， 如果key 里面的连接数为 0 则定期进行清理
	go func() {

		tick := time.NewTicker(time.Millisecond * time.Duration(rl.parm.SyncTick))
		for {
			select {
			case <-tick.C:
				// Synchronize link information
				rl.RLock()

				if rl.electorState == elector.EMaster {
					rl.syncLinkNum()
				}

				rl.RUnlock()
			}
		}

	}()

}

// braid_linker-linknum-gate-base-ukjna1g33rq9
func (rl *redisLinker) getLinkNumKey(child string, id string) string {
	return LinkerRedisPrefix + linknumFlag + splitFlag + rl.serviceName + splitFlag + child + splitFlag + id
}

func (rl *redisLinker) Target(token string, serviceName string) (string, error) {

	rl.RLock()
	defer rl.RUnlock()

	var target string
	var err error

	if rl.parm.Mode == LinkerRedisModeRedis {
		target, err = rl.redisTarget(token, serviceName)
	} else if rl.parm.Mode == LinkerRedisModeLocal {
		target, err = rl.localTarget(token, serviceName)
	}

	return target, err
}

func (rl *redisLinker) Link(token string, target discover.Node) error {
	rl.Lock()
	defer rl.Unlock()

	var err error

	if rl.parm.Mode == LinkerRedisModeRedis {
		err = rl.redisLink(token, target)
	} else if rl.parm.Mode == LinkerRedisModeLocal {
		err = rl.localLink(token, target)
	}

	return err
}

// Unlink 当前节点所属的用户离线
func (rl *redisLinker) Unlink(token string, target string) error {

	rl.Lock()
	defer rl.Unlock()

	var err error

	if rl.parm.Mode == LinkerRedisModeRedis {
		err = rl.redisUnlink(token, target)
	} else if rl.parm.Mode == LinkerRedisModeLocal {
		err = rl.localUnlink(token, target)
	}

	return err
}

// Down 删除离线节点的链路缓存
func (rl *redisLinker) Down(target discover.Node) error {

	rl.Lock()
	defer rl.Unlock()

	var err error

	// need master service
	if rl.electorState != elector.EMaster {
		return nil
	}

	if rl.parm.Mode == LinkerRedisModeRedis {
		err = rl.redisDown(target)
	} else if rl.parm.Mode == LinkerRedisModeLocal {
		err = rl.localDown(target)
	}

	return err
}

func (rl *redisLinker) Close() {
	rl.client.pool.Close()
}

func init() {
	module.Register(newRedisLinker())
}
