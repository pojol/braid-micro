package linkcacheredis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/bredis"
	"github.com/pojol/braid-go/internal/utils"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/service"
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

// Build build link-cache
func BuildWithOption(name string, log *blog.Logger, ps pubsub.IPubsub, client *bredis.Client, opts ...linkcache.Option) linkcache.ILinkCache {

	p := linkcache.Parm{
		Mode:             linkcache.LinkerRedisModeRedis,
		SyncTick:         1000 * 10, // 10 second
		SyncOfflineTick:  60,
		SyncRelationTick: 5,
	}
	for _, opt := range opts {
		opt(&p)
	}

	if client == nil {
		panic(errors.New("linkcache need depend redis client"))
	}

	lc := &redisLinker{
		serviceName:   name,
		electorState:  elector.EWait,
		ps:            ps,
		parm:          p,
		log:           log,
		client:        client,
		activeNodeMap: make(map[string]service.Node),
		local: &localLinker{
			serviceName: name,
			tokenMap:    make(map[string]linkInfo),
			relationSet: make(map[string]int),
		},
	}

	lc.ps.ClusterTopic(name + "." + linkcache.TopicUnlink)
	lc.ps.ClusterTopic(name + "." + linkcache.TopicLinkNum)

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
	parm        linkcache.Parm

	electorState int32
	ps           pubsub.IPubsub
	log          *blog.Logger

	local  *localLinker
	client *bredis.Client

	// 从属节点
	child []string

	activeNodeMap map[string]service.Node

	sync.RWMutex
}

func (rl *redisLinker) Name() string {
	return Name
}

func (rl *redisLinker) Init() error {
	var err error

	ip, err := utils.GetLocalIP()
	if err != nil {
		return fmt.Errorf("%v GetLocalIP err %v", rl.serviceName, err.Error())
	}

	tokenUnlink := rl.ps.ClusterTopic(rl.serviceName + "." + linkcache.TopicUnlink).Sub(Name + "-" + ip)
	serviceUpdate := rl.ps.LocalTopic(discover.TopicServiceUpdate).Sub(Name)
	changeState := rl.ps.LocalTopic(elector.TopicChangeState).Sub(Name)

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
		if statemsg.State != 0 && atomic.LoadInt32(&rl.electorState) != statemsg.State {
			if atomic.CompareAndSwapInt32(&rl.electorState, rl.electorState, statemsg.State) {
				rl.log.Infof("service state change => %v", statemsg.State)
			}
		}
	})

	return nil
}

func (rl *redisLinker) syncLinkNum() {
	conn := rl.getConn()
	defer conn.Close()

	members, err := bredis.ConnSMembers(conn, RelationPrefix)
	if err != nil {
		return
	}

	for _, member := range members {
		info := strings.Split(member, splitFlag)
		if len(info) != 5 {
			rl.log.Warnf("%v wrong relation string format %v", Name, member)
			continue
		}

		parent := info[2]
		id := info[4]
		if rl.serviceName != parent {
			continue
		}

		cnt, err := bredis.ConnGet(conn, member)
		if err != nil {
			rl.log.Warnf("%v redis cmd err %v", Name, err.Error())
			continue
		}

		icnt, err := strconv.Atoi(cnt)
		if err != nil {
			rl.log.Warnf("%v atoi err %v", member, cnt)
		}

		rl.ps.ClusterTopic(rl.serviceName + "." + linkcache.TopicLinkNum).Pub(linkcache.EncodeNumMsg(id, icnt))
	}
}

func (rl *redisLinker) syncRelation() {
	conn := rl.getConn()
	defer conn.Close()

	members, err := bredis.ConnSMembers(conn, RelationPrefix)
	if err != nil {
		return
	}

	rl.Lock()
	defer rl.Unlock()

	childmap := make(map[string]int)

	for _, member := range members {
		info := strings.Split(member, splitFlag)
		if len(info) != 5 {
			rl.log.Warnf("%v wrong relation string format %v", Name, member)
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

func (rl *redisLinker) addOfflineService(service service.Node) {
	rl.Lock()
	rl.activeNodeMap[service.ID] = service
	rl.Unlock()
}

func (rl *redisLinker) rmvOfflineService(service service.Node) {
	rl.Lock()
	delete(rl.activeNodeMap, service.ID)
	rl.Unlock()
}

func (rl *redisLinker) syncOffline() {
	conn := rl.getConn()
	defer conn.Close()

	if rl.parm.Mode != linkcache.LinkerRedisModeLocal && atomic.LoadInt32(&rl.electorState) != elector.EMaster {
		return
	}

	members, err := bredis.ConnSMembers(conn, RelationPrefix)
	if err != nil {
		rl.log.Warnf("smembers %v err %v", RelationPrefix, err.Error())
		return
	}

	rl.Lock()
	defer rl.Unlock()

	offline := []service.Node{}

	for _, member := range members {
		info := strings.Split(member, splitFlag)
		if len(info) != 5 {
			rl.log.Warnf("%v wrong relation string format %v", Name, member)
			continue
		}

		parent := info[2]
		childname := info[3]
		childid := info[4]

		_, ok := rl.activeNodeMap[childid]
		if !ok && rl.serviceName == parent {
			offline = append(offline, service.Node{
				ID:   childid,
				Name: childname,
			})
		}
	}

	for _, service := range offline {
		if rl.parm.Mode == linkcache.LinkerRedisModeLocal {
			err = rl.localDown(service)
		} else if rl.parm.Mode == linkcache.LinkerRedisModeRedis {
			err = rl.redisDown(service)
		}

		rl.log.Debugf("offline service mode:%v, name:%v, id:%v", rl.parm.Mode, service.Name, service.ID)
		if err != nil {
			rl.log.Warnf("offline err %v", err.Error())
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
		tick := time.NewTicker(time.Second * time.Duration(rl.parm.SyncRelationTick))
		for {
			<-tick.C
			rl.syncRelation()
		}
	}()

	go func() {
		tick := time.NewTicker(time.Second * time.Duration(rl.parm.SyncOfflineTick))
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

	if rl.parm.Mode == linkcache.LinkerRedisModeRedis {
		target, err = rl.redisTarget(token, serviceName)
	} else if rl.parm.Mode == linkcache.LinkerRedisModeLocal {
		target, err = rl.localTarget(token, serviceName)
	}

	return target, err
}

func (rl *redisLinker) Link(token string, target service.Node) error {
	rl.Lock()
	defer rl.Unlock()

	var err error

	if rl.parm.Mode == linkcache.LinkerRedisModeRedis {
		err = rl.redisLink(token, target)
	} else if rl.parm.Mode == linkcache.LinkerRedisModeLocal {
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
		if rl.parm.Mode == linkcache.LinkerRedisModeRedis && atomic.LoadInt32(&rl.electorState) == elector.EMaster {
			err = rl.redisUnlink(token, child)
		} else if rl.parm.Mode == linkcache.LinkerRedisModeLocal {
			err = rl.localUnlink(token, child)
		}
	}

	return err
}

// Down 删除离线节点的链路缓存
func (rl *redisLinker) Down(target service.Node) error {

	rl.Lock()
	defer rl.Unlock()

	var err error

	if rl.parm.Mode == linkcache.LinkerRedisModeRedis && atomic.LoadInt32(&rl.electorState) == elector.EMaster {
		err = rl.redisDown(target)
	} else if rl.parm.Mode == linkcache.LinkerRedisModeLocal {
		err = rl.localDown(target)
	}

	return err
}

func (rl *redisLinker) Close() {

}
