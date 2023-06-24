package linkcacheredis

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
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

type linkInfo struct {
	TargetAddr string
	TargetID   string
	TargetName string
}

// redisLinker 基于redis实现的链接器
type redisLinker struct {
	info meta.ServiceInfo
	parm Parm

	electorState int32
	ps           module.IPubsub
	log          *blog.Logger

	local  *localLinker
	client *redis.Client

	tokenUnlink   module.IChannel
	serviceUpdate module.IChannel
	changeState   module.IChannel

	// 从属节点
	child []string

	activeNodeMap map[string]meta.Node

	sync.RWMutex
}

func (rl *redisLinker) Name() string {
	return Name
}

func (rl *redisLinker) Init() error {
	var err error

	rl.tokenUnlink, err = rl.ps.GetTopic(meta.TopicLinkcacheUnlink).
		Sub(context.TODO(), meta.ModuleLink+"-"+rl.info.ID)
	if err != nil {
		return err
	}
	rl.serviceUpdate, err = rl.ps.GetTopic(meta.TopicDiscoverServiceUpdate).
		Sub(context.TODO(), meta.ModuleLink+"-"+rl.info.ID)
	if err != nil {
		return err
	}
	rl.changeState, err = rl.ps.GetTopic(meta.TopicElectionChangeState).
		Sub(context.TODO(), meta.ModuleLink+"-"+rl.info.ID)
	if err != nil {
		return err
	}

	rl.tokenUnlink.Arrived(func(msg *meta.Message) error {
		token := string(msg.Body)
		if token != "" && token != "nil" {
			rl.Unlink(token)
		}
		return nil
	})

	rl.serviceUpdate.Arrived(func(msg *meta.Message) error {
		dmsg := meta.DecodeUpdateMsg(msg)
		if dmsg.Event == meta.TopicDiscoverServiceNodeRmv {
			rl.rmvOfflineService(dmsg.Nod)
			rl.Down(dmsg.Nod)
		} else if dmsg.Event == meta.TopicDiscoverServiceNodeAdd {
			rl.addOfflineService(dmsg.Nod)
		}
		return nil
	})

	rl.changeState.Arrived(func(msg *meta.Message) error {
		statemsg := meta.DecodeStateChangeMsg(msg)
		if statemsg.ID != rl.info.ID {
			return nil
		}

		if statemsg.State != 0 && atomic.LoadInt32(&rl.electorState) != statemsg.State {
			if atomic.CompareAndSwapInt32(&rl.electorState, rl.electorState, statemsg.State) {
				rl.log.Infof("service state change => %v", statemsg.State)
			}
		}
		return nil
	})

	return nil
}

func (rl *redisLinker) syncLinkNum(ctx context.Context) {

	members, err := rl.client.SMembers(ctx, RelationPrefix).Result()
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
		if rl.info.Name != parent {
			continue
		}

		cnt, err := rl.client.Get(ctx, member).Result()
		if err != nil {
			rl.log.Warnf("%v redis cmd err %v", Name, err.Error())
			continue
		}

		icnt, err := strconv.Atoi(cnt)
		if err != nil {
			rl.log.Warnf("%v atoi err %v", member, cnt)
		}

		rl.ps.GetTopic(meta.TopicLinkcacheLinkNumber).Pub(ctx, meta.EncodeNumMsg(id, icnt))
	}
}

func (rl *redisLinker) syncRelation(ctx context.Context) {

	members, err := rl.client.SMembers(ctx, RelationPrefix).Result()
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
		if parent == rl.info.Name {
			childmap[child] = 1
		}
	}

	rl.child = rl.child[:0]
	for newchild := range childmap {
		rl.child = append(rl.child, newchild)
	}
}

func (rl *redisLinker) addOfflineService(service meta.Node) {
	rl.Lock()
	rl.activeNodeMap[service.ID] = service
	rl.Unlock()
}

func (rl *redisLinker) rmvOfflineService(service meta.Node) {
	rl.Lock()
	delete(rl.activeNodeMap, service.ID)
	rl.Unlock()
}

func (rl *redisLinker) syncOffline(ctx context.Context) {

	if rl.parm.Mode != LinkerRedisModeLocal && atomic.LoadInt32(&rl.electorState) != meta.EMaster {
		return
	}

	members, err := rl.client.SMembers(ctx, RelationPrefix).Result()
	if err != nil {
		rl.log.Warnf("smembers %v err %v", RelationPrefix, err.Error())
		return
	}

	rl.Lock()
	defer rl.Unlock()

	offline := []meta.Node{}

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
		if !ok && rl.info.Name == parent {
			offline = append(offline, meta.Node{
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

	rl.syncRelation(context.TODO())
	go func() {
		tick := time.NewTicker(time.Second * time.Duration(rl.parm.SyncRelationTick))
		for {
			<-tick.C
			rl.syncRelation(context.TODO())
		}
	}()

	go func() {
		tick := time.NewTicker(time.Second * time.Duration(rl.parm.SyncOfflineTick))
		for {
			<-tick.C
			rl.syncOffline(context.TODO())
		}
	}()
}

// braid_linker-linknum-gate-base-ukjna1g33rq9
func (rl *redisLinker) getLinkNumKey(child string, id string) string {
	return LinkerRedisPrefix + linknumFlag + splitFlag + rl.info.Name + splitFlag + child + splitFlag + id
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

func (rl *redisLinker) Link(token string, target meta.Node) error {
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
		if rl.parm.Mode == LinkerRedisModeRedis && atomic.LoadInt32(&rl.electorState) == meta.EMaster {
			err = rl.redisUnlink(token, child)
		} else if rl.parm.Mode == LinkerRedisModeLocal {
			err = rl.localUnlink(token, child)
		}
	}

	return err
}

// Down 删除离线节点的链路缓存
func (rl *redisLinker) Down(target meta.Node) error {

	rl.Lock()
	defer rl.Unlock()

	var err error

	if rl.parm.Mode == LinkerRedisModeRedis && atomic.LoadInt32(&rl.electorState) == meta.EMaster {
		err = rl.redisDown(target)
	} else if rl.parm.Mode == LinkerRedisModeLocal {
		err = rl.localDown(target)
	}

	return err
}

func (rl *redisLinker) Close() {
	rl.changeState.Close()
	rl.serviceUpdate.Close()
	rl.tokenUnlink.Close()
}
