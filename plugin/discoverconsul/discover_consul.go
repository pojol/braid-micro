package discoverconsul

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pojol/braid/3rd/consul"
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/pubsub"
)

const (
	// DiscoverName 发现器名称
	DiscoverName = "ConsulDiscover"

	// DiscoverTag 用于docker发现的tag， 所有希望被discover服务发现的节点，
	// 都应该在Dockerfile中设置 ENV SERVICE_TAGS=braid
	DiscoverTag = "braid"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")

	// 权重预设值，可以约等于节点支持的最大连接数
	// 在开启linker的情况下，节点的连接数越多权重值就越低，直到降到最低的 1权重
	defaultWeight = 1000
)

type consulDiscoverBuilder struct {
	cfg Cfg
}

func newConsulDiscover() discover.Builder {
	return &consulDiscoverBuilder{}
}

func (b *consulDiscoverBuilder) Name() string {
	return DiscoverName
}

func (b *consulDiscoverBuilder) SetCfg(cfg interface{}) error {
	cecfg, ok := cfg.(Cfg)
	if !ok {
		return ErrConfigConvert
	}

	b.cfg = cecfg
	if b.cfg.Tag == "" {
		b.cfg.Tag = strings.ToLower(DiscoverTag)
	}
	return nil
}

func (b *consulDiscoverBuilder) Build(ps pubsub.IPubsub, linker linkcache.ILinkCache) discover.IDiscover {

	e := &consulDiscover{
		cfg:        b.cfg,
		pubsub:     ps,
		linker:     linker,
		passingMap: make(map[string]*syncNode),
	}

	return e
}

// Discover 发现管理braid相关的节点
type consulDiscover struct {
	discoverTicker   *time.Ticker
	syncWeightTicker *time.Ticker

	cfg    Cfg
	pubsub pubsub.IPubsub
	linker linkcache.ILinkCache

	// service id : service nod
	passingMap map[string]*syncNode
}

type syncNode struct {
	service string
	id      string
	address string

	dyncWeight int
	physWeight int
}

func (dc *consulDiscover) InBlacklist(name string) bool {

	for _, v := range dc.cfg.Blacklist {
		if v == name {
			return true
		}
	}

	return false
}

func (dc *consulDiscover) discoverImpl() {

	services, err := consul.GetCatalogServices(dc.cfg.Address, dc.cfg.Tag)
	if err != nil {
		return
	}

	for _, service := range services {
		if service.ServiceName == dc.cfg.Name {
			continue
		}

		if dc.InBlacklist(service.ServiceName) {
			continue
		}

		if service.ServiceName == "" || service.ServiceID == "" {
			continue
		}

		if _, ok := dc.passingMap[service.ServiceID]; !ok { // new nod
			// regist service
			balancer.Get(service.ServiceName)
			log.Debugf("new service %s id %s", service.ServiceName, service.ID)

			sn := syncNode{
				service:    service.ServiceName,
				id:         service.ServiceID,
				address:    service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort),
				dyncWeight: 0,
				physWeight: defaultWeight,
			}

			dc.passingMap[service.ServiceID] = &sn

			dc.pubsub.Pub(discover.EventAdd+"_"+service.ServiceName, pubsub.NewMessage(discover.Node{
				ID:      sn.id,
				Name:    sn.service,
				Address: sn.address,
				Weight:  sn.physWeight,
			}))

		}
	}

	for k := range dc.passingMap {
		if _, ok := services[k]; !ok { // rmv nod
			log.Debugf("remove service %s id %s", dc.passingMap[k].service, dc.passingMap[k].id)

			dc.pubsub.Pub(discover.EventRmv+"_"+dc.passingMap[k].service, pubsub.NewMessage(discover.Node{
				ID:   dc.passingMap[k].id,
				Name: dc.passingMap[k].service,
			}))

			if dc.linker != nil {
				dc.linker.Down(dc.passingMap[k].service, dc.passingMap[k].address)
			}

			delete(dc.passingMap, k)
		}
	}
}

func (dc *consulDiscover) syncWeight() {
	for k, v := range dc.passingMap {
		num, err := dc.linker.Num(v.service, v.address)
		if err != nil || num == 0 {
			continue
		}

		if num == v.dyncWeight {
			continue
		}

		dc.passingMap[k].dyncWeight = num
		nweight := 0
		if dc.passingMap[k].physWeight-num > 0 {
			nweight = dc.passingMap[k].physWeight - num
		} else {
			nweight = 1
		}

		dc.pubsub.Pub(discover.EventUpdate+"_"+v.service, pubsub.NewMessage(discover.Node{
			ID:     v.id,
			Name:   v.service,
			Weight: nweight,
		}))
	}
}

func (dc *consulDiscover) runImpl() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("status", "sync service", fmt.Errorf("%v", err).Error())
			}
		}()
		// todo ..
		dc.discoverImpl()
	}

	syncWeight := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("status", "sync weight", fmt.Errorf("%v", err).Error())
			}
		}()

		if dc.linker != nil {
			dc.syncWeight()
		}

	}

	dc.discoverTicker = time.NewTicker(dc.cfg.Interval)
	dc.syncWeightTicker = time.NewTicker(time.Second * 10)

	dc.discoverImpl()

	for {
		select {
		case <-dc.discoverTicker.C:
			syncService()
		case <-dc.syncWeightTicker.C:
			syncWeight()
		}
	}
}

// Discover 运行管理器
func (dc *consulDiscover) Discover() {
	go func() {
		dc.runImpl()
	}()
}

// Close close
func (dc *consulDiscover) Close() {

}

func init() {
	discover.Register(newConsulDiscover())
}
