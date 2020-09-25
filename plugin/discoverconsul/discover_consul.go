package discoverconsul

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pojol/braid/3rd/consul"
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/pubsub"
	"github.com/pojol/braid/plugin/linkerredis"
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
	defaultWeight = 1000
)

type consulDiscoverBuilder struct {
	opts []interface{}
}

func newConsulDiscover() discover.Builder {
	return &consulDiscoverBuilder{}
}

func (b *consulDiscoverBuilder) Name() string {
	return Name
}

func (b *consulDiscoverBuilder) AddOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *consulDiscoverBuilder) Build(serviceName string) (discover.IDiscover, error) {

	p := Parm{
		Tag:      "braid",
		Name:     serviceName,
		Interval: time.Second * 2,
		Address:  "http://127.0.0.1:8500",
	}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	if p.procPB == nil {
		return nil, errors.New("discover_consul parm mismatch " + "no proc pubsub!")
	}

	// check address
	_, err := consul.GetCatalogServices(p.Address, p.Tag)
	fmt.Println(p.Address, err)
	if err != nil {
		return nil, err
	}

	e := &consulDiscover{
		parm:       p,
		passingMap: make(map[string]*syncNode),
	}

	return e, nil
}

// Discover 发现管理braid相关的节点
type consulDiscover struct {
	discoverTicker   *time.Ticker
	syncWeightTicker *time.Ticker

	// parm
	parm Parm

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

	for _, v := range dc.parm.Blacklist {
		if v == name {
			return true
		}
	}

	return false
}

func (dc *consulDiscover) discoverImpl() {

	services, err := consul.GetCatalogServices(dc.parm.Address, dc.parm.Tag)
	if err != nil {
		return
	}

	for _, service := range services {
		if service.ServiceName == dc.parm.Name {
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

			sn := syncNode{
				service:    service.ServiceName,
				id:         service.ServiceID,
				address:    service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort),
				dyncWeight: 0,
				physWeight: defaultWeight,
			}
			log.Debugf("new service %s addr %s", service.ServiceName, sn.address)

			dc.passingMap[service.ServiceID] = &sn

			dc.parm.procPB.Pub(discover.EventAdd+"_"+service.ServiceName, pubsub.NewMessage(discover.Node{
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

			dc.parm.procPB.Pub(discover.EventRmv+"_"+dc.passingMap[k].service, pubsub.NewMessage(discover.Node{
				ID:   dc.passingMap[k].id,
				Name: dc.passingMap[k].service,
			}))

			if dc.parm.clusterPB != nil {
				dc.parm.clusterPB.Pub(linkerredis.LinkerTopicDown, linkerredis.NewDownMsg(
					dc.passingMap[k].service,
					dc.passingMap[k].address,
				))
			}

			delete(dc.passingMap, k)
		}
	}
}

func (dc *consulDiscover) syncWeight() {
	for k, v := range dc.passingMap {
		num, err := dc.parm.linkcache.Num(discover.Node{
			ID:      v.id,
			Name:    v.service,
			Address: v.address,
		})
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

		dc.parm.procPB.Pub(discover.EventUpdate+"_"+v.service, pubsub.NewMessage(discover.Node{
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

		if dc.parm.linkcache != nil {
			dc.syncWeight()
		}

	}

	dc.discoverTicker = time.NewTicker(dc.parm.Interval)
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
