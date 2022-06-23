// 实现文件 基于 consul 实现的服务发现
package discover

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/depend/pubsub"
	"github.com/pojol/braid-go/internal/utils"
	"github.com/pojol/braid-go/service"
)

const (
	// Name 发现器名称
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

func Build(name string, opts ...Option) IDiscover {

	p := Parm{
		Tag:                       "braid",
		Name:                      name,
		SyncServicesInterval:      time.Second * 2,
		SyncServiceWeightInterval: time.Second * 10,
		Address:                   "http://127.0.0.1:8500",
	}

	for _, opt := range opts {
		opt(&p)
	}

	e := &consulDiscover{
		parm:       p,
		passingMap: make(map[string]*syncNode),
	}

	e.ps.GetTopic(service.TopicServiceUpdate)

	return e
}

func (dc *consulDiscover) Init() error {

	// check address
	_, err := consul.GetConsulLeader(dc.parm.Address)
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", dc.parm.Name, "consul", dc.parm.Address)
	}

	ip, err := utils.GetLocalIP()
	if err != nil {
		return fmt.Errorf("%v GetLocalIP err %v", dc.parm.Name, err.Error())
	}

	linkC := dc.ps.GetTopic(service.TopicLinkerLinkNum).Sub(Name + "-" + ip)
	linkC.Arrived(func(msg *pubsub.Message) {
		lninfo := service.LinkerDecodeNumMsg(msg)
		dc.lock.Lock()
		defer dc.lock.Unlock()

		if _, ok := dc.passingMap[lninfo.ID]; ok {
			dc.passingMap[lninfo.ID].linknum = lninfo.Num
		}
	})

	return nil
}

// Discover 发现管理braid相关的节点
type consulDiscover struct {
	discoverTicker   *time.Ticker
	syncWeightTicker *time.Ticker

	// parm
	parm Parm
	ps   pubsub.IPubsub

	// service id : service nod
	passingMap map[string]*syncNode

	lock sync.Mutex
}

type syncNode struct {
	service string
	id      string
	address string

	linknum int

	dyncWeight int
	physWeight int
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

	services, err := consul.GetCatalogServices(dc.parm.Address, dc.parm.Tag)
	if err != nil {
		return
	}

	for _, cs := range services {
		if cs.ServiceName == dc.parm.Name {
			continue
		}

		if dc.InBlacklist(cs.ServiceName) {
			continue
		}

		if cs.ServiceName == "" || cs.ServiceID == "" {
			continue
		}

		if _, ok := dc.passingMap[cs.ServiceID]; !ok { // new nod
			sn := syncNode{
				service:    cs.ServiceName,
				id:         cs.ServiceID,
				address:    cs.ServiceAddress + ":" + strconv.Itoa(cs.ServicePort),
				dyncWeight: 0,
				physWeight: defaultWeight,
			}
			blog.Infof("new service %s addr %s", cs.ServiceName, sn.address)
			dc.passingMap[cs.ServiceID] = &sn

			dc.ps.GetTopic(service.TopicServiceUpdate).Pub(service.DiscoverEncodeUpdateMsg(
				EventAddService,
				service.Node{
					ID:      sn.id,
					Name:    sn.service,
					Address: sn.address,
					Weight:  sn.physWeight,
				},
			))
		}
	}

	for k := range dc.passingMap {
		if _, ok := services[k]; !ok { // rmv nod
			blog.Infof("remove service %s id %s", dc.passingMap[k].service, dc.passingMap[k].id)

			dc.ps.GetTopic(service.TopicServiceUpdate).Pub(service.DiscoverEncodeUpdateMsg(
				EventRemoveService,
				service.Node{
					ID:      dc.passingMap[k].id,
					Name:    dc.passingMap[k].service,
					Address: dc.passingMap[k].address,
				},
			))

			delete(dc.passingMap, k)
		}
	}
}

func (dc *consulDiscover) syncWeight() {
	dc.lock.Lock()
	defer dc.lock.Unlock()

	for k, v := range dc.passingMap {
		if v.linknum == 0 {
			continue
		}

		if v.linknum == v.dyncWeight {
			continue
		}

		dc.passingMap[k].dyncWeight = v.linknum
		nweight := 0
		if dc.passingMap[k].physWeight-v.linknum > 0 {
			nweight = dc.passingMap[k].physWeight - v.linknum
		} else {
			nweight = 1
		}

		dc.ps.GetTopic(service.TopicServiceUpdate).Pub(service.DiscoverEncodeUpdateMsg(
			EventUpdateService,
			service.Node{
				ID:     v.id,
				Name:   v.service,
				Weight: nweight,
			},
		))
	}
}

func (dc *consulDiscover) discover() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				blog.Errf("consul discover syncService err %v", err)
			}
		}()
		// todo ..
		dc.discoverImpl()
	}

	dc.discoverTicker = time.NewTicker(dc.parm.SyncServicesInterval)

	dc.discoverImpl()

	for {
		select {
		case <-dc.discoverTicker.C:
			syncService()
		}
	}
}

func (dc *consulDiscover) weight() {
	syncWeight := func() {
		defer func() {
			if err := recover(); err != nil {
				blog.Errf("consul discover syncWeight err %v", err)
			}
		}()

		dc.syncWeight()
	}

	dc.syncWeightTicker = time.NewTicker(dc.parm.SyncServiceWeightInterval)

	for {
		select {
		case <-dc.syncWeightTicker.C:
			syncWeight()
		}
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
