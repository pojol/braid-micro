// 实现文件 基于 consul 实现的服务发现
package discoverconsul

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/pojol/braid-go/depend/bconsul"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/internal/utils"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/service"
)

const (
	// Name 服务发现
	Name = "ConsulDiscover"

	// DiscoverTag 用于docker发现的tag， 所有希望被discover服务发现的节点，
	// 都应该在Dockerfile中设置 ENV SERVICE_TAGS=braid
	DiscoverTag = "braid"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")

	// 权重预设值，可以约等于节点支持的最大连接数
	// 在开启linker的情况下，节点的连接数越多权重值就越低，直到降到最低的 1权重
	defaultWeight = 1024
)

func BuildWithOption(name string, log *blog.Logger, ps pubsub.IPubsub, client *bconsul.Client, opts ...discover.Option) discover.IDiscover {

	p := discover.Parm{
		Tag:                       "braid",
		Name:                      name,
		SyncServicesInterval:      time.Second * 2,
		SyncServiceWeightInterval: time.Second * 10,
	}

	for _, opt := range opts {
		opt(&p)
	}

	if client == nil {
		panic(errors.New("discover need depend consul client"))
	}

	e := &consulDiscover{
		parm:    p,
		ps:      ps,
		log:     log,
		client:  client,
		nodemap: make(map[string]*service.Node),
	}

	e.ps.Topic(discover.TopicServiceUpdate)

	return e
}

func (dc *consulDiscover) Name() string {
	return Name
}

func (dc *consulDiscover) Init() error {
	/*
		ip, err := utils.GetLocalIP()
		if err != nil {
			return fmt.Errorf("%v GetLocalIP err %v", dc.parm.Name, err.Error())
		}

			linkC := dc.ps.GetTopic(service.TopicLinkerLinkNum).Sub(Name + "-" + ip)
			linkC.Arrived(func(msg *pubsub.Message) {
				lninfo := service.LinkerDecodeNumMsg(msg)
				dc.lock.Lock()
				defer dc.lock.Unlock()

				if _, ok := dc.nodemap[lninfo.ID]; ok {
					dc.nodemap[lninfo.ID].linknum = lninfo.Num
				}
			})
	*/
	return nil
}

// Discover 发现管理braid相关的节点
type consulDiscover struct {
	discoverTicker   *time.Ticker
	syncWeightTicker *time.Ticker

	client *bconsul.Client

	// parm
	parm discover.Parm
	ps   pubsub.IPubsub
	log  *blog.Logger

	// service id : service nod
	nodemap map[string]*service.Node

	lock sync.Mutex
}

func (dc *consulDiscover) InBlacklist(name string) bool {

	for _, v := range dc.parm.Blacklist {
		if v == name {
			return true
		}
	}

	return false
}

func (dc *consulDiscover) discoverImpl() {

	dc.lock.Lock()
	defer dc.lock.Unlock()

	servicesnodes := make(map[string]bool)

	services, err := dc.client.CatalogListServices()
	if err != nil {
		dc.log.Warnf("[Discover] discover impl err %v", err.Error())
		return
	}

	for _, v := range services {
		cs, err := dc.client.CatalogGetService(v.Name)
		if err != nil {
			dc.log.Warnf("[Discover] catalog get service err %v", err)
			continue
		}

		if v.Name == "" || len(cs.Nodes) == 0 {
			continue
		}

		if !utils.ContainsInSlice(v.Tags, dc.parm.Tag) {
			dc.log.Debugf("[Discover] rule out with service tag %v, self tag %v", v.Tags, dc.parm.Tag)
			continue
		}

		if v.Name == dc.parm.Name {
			dc.log.Debugf("[Discover] rule out with self")
			continue
		}

		if utils.ContainsInSlice(dc.parm.Blacklist, v.Name) {
			dc.log.Debugf("[Discover] rule out with black list %v", v.Name)
			continue // 排除黑名单节点
		}

		// 添加节点
		for _, nod := range cs.Nodes {

			servicesnodes[nod.ID] = true

			if _, ok := dc.nodemap[nod.ID]; !ok {

				sn := service.Node{
					Name:    v.Name,
					ID:      nod.ID,
					Address: nod.Address + ":" + strconv.Itoa(nod.Port),
					Weight:  defaultWeight,
				}
				dc.log.Infof("[Discover] new service %s node %s addr %s", v.Name, nod.ID, sn.Address)
				dc.nodemap[nod.ID] = &sn

				dc.ps.Topic(discover.TopicServiceUpdate).Pub(context.TODO(), discover.EncodeUpdateMsg(
					discover.EventAddService,
					sn,
				))

			}

		}

	}

	// 排除节点
	for k := range dc.nodemap {

		if _, ok := servicesnodes[k]; !ok {
			dc.log.Infof("[Discover] remove service %s node %s", dc.nodemap[k].Name, dc.nodemap[k].ID)

			dc.ps.Topic(discover.TopicServiceUpdate).Pub(context.TODO(), discover.EncodeUpdateMsg(
				discover.EventRemoveService,
				*dc.nodemap[k],
			))

			delete(dc.nodemap, k)
		}

	}

}

func (dc *consulDiscover) syncWeight() {
	dc.lock.Lock()
	defer dc.lock.Unlock()

	/*
		for k, v := range dc.nodemap {
			if v.linknum == 0 {
				continue
			}

			if v.linknum == v.dyncWeight {
				continue
			}

			dc.nodemap[k].dyncWeight = v.linknum
			nweight := 0
			if dc.nodemap[k].physWeight-v.linknum > 0 {
				nweight = dc.nodemap[k].physWeight - v.linknum
			} else {
				nweight = 1
			}

			dc.ps.LocalTopic(discover.TopicServiceUpdate).Pub(discover.EncodeUpdateMsg(
				discover.EventUpdateService,
				service.Node{
					ID:     v.id,
					Name:   v.service,
					Weight: nweight,
				},
			))
		}
	*/
}

func (dc *consulDiscover) discover() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				dc.log.Errf("[Discover] syncService err %v", err)
			}
		}()
		// todo ..
		dc.discoverImpl()
	}

	dc.discoverTicker = time.NewTicker(dc.parm.SyncServicesInterval)

	dc.discoverImpl()

	for {
		<-dc.discoverTicker.C
		syncService()
	}
}

func (dc *consulDiscover) weight() {
	syncWeight := func() {
		defer func() {
			if err := recover(); err != nil {
				dc.log.Errf("[Discover] syncWeight err %v", err)
			}
		}()

		dc.syncWeight()
	}

	dc.syncWeightTicker = time.NewTicker(dc.parm.SyncServiceWeightInterval)

	for {
		<-dc.syncWeightTicker.C
		syncWeight()
	}
}

// Discover 运行管理器
func (dc *consulDiscover) Run() {
	go func() {
		dc.discover()
	}()

	go func() {
		dc.weight()
	}()
}

// Close close
func (dc *consulDiscover) Close() {

}
