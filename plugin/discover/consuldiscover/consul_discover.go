package consuldiscover

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pojol/braid/3rd/consul"
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/plugin/balancer"
	"github.com/pojol/braid/plugin/discover"
	"github.com/pojol/braid/plugin/linker"
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
)

type consulDiscoverBuilder struct{}

func newConsulDiscover() discover.Builder {
	return &consulDiscoverBuilder{}
}

func (*consulDiscoverBuilder) Name() string {
	return DiscoverName
}

func (*consulDiscoverBuilder) Build(bg *balancer.Group, linker linker.ILinker, cfg interface{}) discover.IDiscover {
	cecfg, ok := cfg.(Cfg)
	if !ok {
		return nil
	}

	e := &consulDiscover{
		cfg:        cecfg,
		bg:         bg,
		passingMap: make(map[string]syncNode),
		linker:     linker,
	}

	return e
}

// Cfg discover config
type Cfg struct {
	Name string

	// 同步节点信息间隔
	Interval time.Duration

	// 注册中心
	ConsulAddress string
}

// Discover 发现管理braid相关的节点
type consulDiscover struct {
	discoverTicker   *time.Ticker
	syncWeightTicker *time.Ticker

	cfg Cfg
	bg  *balancer.Group

	linker linker.ILinker

	// service id : service nod
	passingMap map[string]syncNode
}

type syncNode struct {
	service string
	id      string
	address string
}

func (dc *consulDiscover) discoverImpl() {

	services, err := consul.GetCatalogServices(dc.cfg.ConsulAddress, DiscoverTag)
	if err != nil {
		return
	}

	for _, service := range services {
		if service.ServiceName == dc.cfg.Name {
			continue
		}

		if _, ok := dc.passingMap[service.ServiceID]; !ok { // new nod

			sn := syncNode{
				service: service.ServiceName,
				id:      service.ServiceID,
				address: service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort),
			}

			dc.passingMap[service.ServiceID] = sn
			dc.bg.Get(sn.service).Update(balancer.Node{
				ID:      sn.id,
				Name:    sn.service,
				Address: sn.address,
				Weight:  1,
				OpTag:   balancer.OpAdd,
			})

		}
	}

	for k := range dc.passingMap {
		if _, ok := services[k]; !ok { // rmv nod

			dc.bg.Get(dc.passingMap[k].service).Update(balancer.Node{
				ID:    dc.passingMap[k].id,
				Name:  dc.passingMap[k].service,
				OpTag: balancer.OpRmv,
			})

			if dc.linker != nil {
				dc.linker.Offline(dc.passingMap[k].id)
			}

			delete(dc.passingMap, k)
		}
	}
}

func (dc *consulDiscover) SyncWeight() {
	if dc.linker == nil {
		return
	}

	for k := range dc.passingMap {
		num, err := dc.linker.Num(dc.passingMap[k].id)
		if err != nil {
			continue
		}

		dc.bg.Get(dc.passingMap[k].service).Update(balancer.Node{
			ID:      dc.passingMap[k].id,
			Name:    dc.passingMap[k].service,
			Address: dc.passingMap[k].address,
			Weight:  num,
			OpTag:   balancer.OpUp,
		})
	}

}

func (dc *consulDiscover) runImpl() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("status", "sync", fmt.Errorf("%v", err).Error())
			}
		}()
		// todo ..
		dc.discoverImpl()
	}

	dc.discoverTicker = time.NewTicker(dc.cfg.Interval)
	dc.syncWeightTicker = time.NewTicker(time.Second * 10)
	dc.discoverImpl()

	for {
		select {
		case <-dc.discoverTicker.C:
			syncService()
		case <-dc.syncWeightTicker.C:
			dc.SyncWeight()
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
