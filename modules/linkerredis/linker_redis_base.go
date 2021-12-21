package linkerredis

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pojol/braid-go/internal/utils"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/moduleparm"
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

func newRedisLinker() module.IBuilder {
	return &redisLinkerBuilder{}
}

func (*redisLinkerBuilder) Name() string {
	return Name
}

func (*redisLinkerBuilder) Type() module.ModuleType {
	return module.Linkcache
}

func (rb *redisLinkerBuilder) AddModuleOption(opt interface{}) {
	rb.opts = append(rb.opts, opt)
}

// Build build link-cache
func (rb *redisLinkerBuilder) Build(name string, buildOpts ...interface{}) interface{} {

	bp := moduleparm.BuildParm{}
	for _, opt := range buildOpts {
		opt.(moduleparm.Option)(&bp)
	}

	p := Parm{
		Mode:             LinkerRedisModeRedis,
		SyncTick:         1000 * 10, // 10 second
		RedisAddr:        "redis://127.0.0.1:6379/0",
		RedisMaxIdle:     16,
		RedisMaxActive:   128,
		syncOfflineTick:  60,
		syncRelationTick: 5,
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
		serviceName:   name,
		ps:            bp.PS,
		electorState:  elector.EWait,
		logger:        bp.Logger,
		client:        client,
		parm:          p,
		activeNodeMap: make(map[string]discover.Node),
		local: &localLinker{
			serviceName: name,
			tokenMap:    make(map[string]linkInfo),
			relationSet: make(map[string]int),
		},
	}

	lc.ps.RegistTopic(linkcache.TokenUnlink, pubsub.ScopeCluster)
	lc.ps.RegistTopic(linkcache.ServiceLinkNum, pubsub.ScopeCluster)

	return lc
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
	ps           pubsub.IPubsub

	logger logger.ILogger

	client *RedisClient
	local  *localLinker

	// 从属节点
	child []string

	activeNodeMap map[string]discover.Node

	sync.RWMutex
}

func (rl *redisLinker) Init() error {
	var err error

	ip, err := utils.GetLocalIP()
	if err != nil {
		return fmt.Errorf("%v GetLocalIP err %v", rl.serviceName, err.Error())
	}

	tokenUnlink := rl.ps.GetTopic(linkcache.TokenUnlink).Sub(Name + "-" + ip)
	serviceUpdate := rl.ps.GetTopic(discover.ServiceUpdate).Sub(Name)
	changeState := rl.ps.GetTopic(elector.ChangeState).Sub(Name)

	tokenUnlink.Arrived(func(msg *pubsub.Message) {
		token := string(msg.Body)
		if token != "" && token != "nil" {
			rl.Unlink(token)
		}
	})

	serviceUpdate.Arrived(func(msg *pubsub.Message) {
		dmsg := discover.DecodeUpdateMsg(msg)
		if dmsg.Event == discover.EventRemoveService {
			rl.rmvOfflineService(dmsg.Nod)
			rl.Down(dmsg.Nod)
		} else if dmsg.Event == discover.EventAddService {
			rl.addOfflineService(dmsg.Nod)
		}
	})

	changeState.Arrived(func(msg *pubsub.Message) {
		statemsg := elector.DecodeStateChangeMsg(msg)
		if statemsg.State != "" && rl.electorState != statemsg.State {
			rl.electorState = statemsg.State
			rl.logger.Debugf("service state change => %v", statemsg.State)
		}
	})

	_, err = rl.client.Ping()
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", rl.serviceName, "redis", rl.parm.RedisAddr)
	}

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

		rl.ps.GetTopic(linkcache.ServiceLinkNum).Pub(linkcache.EncodeLinkNumMsg(id, int(cnt)))
	}
}

func (rl *redisLinker) syncRelation() {
	conn := rl.getConn()
	defer conn.Close()

	members, err := redis.Strings(conn.Do("SMEMBERS", RelationPrefix))
	if err != nil {
		return
	}

	rl.Lock()
	defer rl.Unlock()

	childmap := make(map[string]int)

	for _, member := range members {
		info := strings.Split(member, splitFlag)
		if len(info) != 5 {
			rl.logger.Warnf("%v wrong relation string format %v", Name, member)
			continue
		}

		parent := info[2]
		child := info[3]
		if parent == rl.serviceName {
			childmap[child] = 1
		}
	}

	rl.child = rl.child[:0]
	for newchild := range childmap {
		rl.child = append(rl.child, newchild)
	}
}

func (rl *redisLinker) addOfflineService(service discover.Node) {
	rl.Lock()
	rl.activeNodeMap[service.ID] = service
	rl.Unlock()
}

func (rl *redisLinker) rmvOfflineService(service discover.Node) {
	rl.Lock()
	delete(rl.activeNodeMap, service.ID)
	rl.Unlock()
}

func (rl *redisLinker) syncOffline() {
	conn := rl.getConn()
	defer conn.Close()

	if rl.parm.Mode != LinkerRedisModeLocal && rl.electorState != elector.EMaster {
		return
	}

	members, err := redis.Strings(conn.Do("SMEMBERS", RelationPrefix))
	if err != nil {
		rl.logger.Warnf("smembers %v err %v", RelationPrefix, err.Error())
		return
	}

	rl.Lock()
	defer rl.Unlock()

	offline := []discover.Node{}

	for _, member := range members {
		info := strings.Split(member, splitFlag)
		if len(info) != 5 {
			rl.logger.Warnf("%v wrong relation string format %v", Name, member)
			continue
		}

		parent := info[2]
		childname := info[3]
		childid := info[4]

		_, ok := rl.activeNodeMap[childid]
		if !ok && rl.serviceName == parent {
			offline = append(offline, discover.Node{
				ID:   childid,
				Name: childname,
			})
		}
	}

	for _, service := range offline {
		if rl.parm.Mode == LinkerRedisModeLocal {
			err = rl.localDown(service)
		} else if rl.parm.Mode == LinkerRedisModeRedis {
			err = rl.redisDown(service)
		}

		rl.logger.Debugf("offline service mode:%v, name:%v, id:%v", rl.parm.Mode, service.Name, service.ID)
		if err != nil {
			rl.logger.Warnf("offline err %v", err.Error())
		}
	}
}

func (rl *redisLinker) Run() {

	/*
		// 暂时屏蔽这段代码，因为在swarm模式下，没有办法设置物理权重；因此暂不调整权重。
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
	*/

	rl.syncRelation()
	go func() {
		tick := time.NewTicker(time.Second * time.Duration(rl.parm.syncRelationTick))
		for {
			<-tick.C
			rl.syncRelation()
		}
	}()

	go func() {
		tick := time.NewTicker(time.Second * time.Duration(rl.parm.syncOfflineTick))
		for {
			<-tick.C
			rl.syncOffline()
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
func (rl *redisLinker) Unlink(token string) error {

	rl.Lock()
	defer rl.Unlock()

	var err error

	// 尝试将自身名下的节点中的token释放掉
	for _, child := range rl.child {
		if rl.parm.Mode == LinkerRedisModeRedis && rl.electorState == elector.EMaster {
			err = rl.redisUnlink(token, child)
		} else if rl.parm.Mode == LinkerRedisModeLocal {
			err = rl.localUnlink(token, child)
		}
	}

	return err
}

// Down 删除离线节点的链路缓存
func (rl *redisLinker) Down(target discover.Node) error {

	rl.Lock()
	defer rl.Unlock()

	var err error

	if rl.parm.Mode == LinkerRedisModeRedis && rl.electorState == elector.EMaster {
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
