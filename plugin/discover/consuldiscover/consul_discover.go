package consuldiscover

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pojol/braid/3rd/consul"
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/plugin/balancer"
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

func (b *consulDiscoverBuilder) Build() discover.IDiscover {

	e := &consulDiscover{
		cfg:        b.cfg,
		passingMap: make(map[string]*syncNode),
	}

	return e
}

// Cfg discover config
type Cfg struct {
	Name string

	// 同步节点信息间隔
	Interval time.Duration

	// 注册中心
	Address string

	Tag string
}

// Option consul discover config wrapper
type Option func(*Cfg)

// WithTag 修改config中的discover tag
func WithTag(discoverTag string) Option {
	return func(c *Cfg) {
		c.Tag = discoverTag
	}
}

// WithInterval 修改config中的interval
func WithInterval(interval time.Duration) Option {
	return func(c *Cfg) {
		c.Interval = interval
	}
}

// Discover 发现管理braid相关的节点
type consulDiscover struct {
	discoverTicker   *time.Ticker
	syncWeightTicker *time.Ticker

	cfg Cfg

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

func (dc *consulDiscover) discoverImpl() {

	services, err := consul.GetCatalogServices(dc.cfg.Address, dc.cfg.Tag)
	if err != nil {
		return
	}

	for _, service := range services {
		if service.ServiceName == dc.cfg.Name {
			continue
		}

		if _, ok := dc.passingMap[service.ServiceID]; !ok { // new nod

			sn := syncNode{
				service:    service.ServiceName,
				id:         service.ServiceID,
				address:    service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort),
				dyncWeight: 0,
				physWeight: defaultWeight,
			}

			dc.passingMap[service.ServiceID] = &sn
			balancer.Get(sn.service).Update(balancer.Node{
				ID:      sn.id,
				Name:    sn.service,
				Address: sn.address,
				Weight:  sn.physWeight,
				OpTag:   balancer.OpAdd,
			})

		}
	}

	for k := range dc.passingMap {
		if _, ok := services[k]; !ok { // rmv nod

			balancer.Get(dc.passingMap[k].service).Update(balancer.Node{
				ID:    dc.passingMap[k].id,
				Name:  dc.passingMap[k].service,
				OpTag: balancer.OpRmv,
			})

			delete(dc.passingMap, k)
		}
	}
}

/*
func (dc *consulDiscover) SyncWeight() {

	for k := range dc.passingMap {
		num, err := dc.linker.Num(dc.passingMap[k].id)
		if err != nil || num == 0 {
			continue
		}

		if num == dc.passingMap[k].dyncWeight {
			continue
		}

		dc.passingMap[k].dyncWeight = num
		nweight := 0
		if dc.passingMap[k].physWeight-num > 0 {
			nweight = dc.passingMap[k].physWeight - num
		} else {
			nweight = 1
		}

		balancer.Get(dc.passingMap[k].service).Update(balancer.Node{
			ID:      dc.passingMap[k].id,
			Name:    dc.passingMap[k].service,
			Address: dc.passingMap[k].address,
			Weight:  nweight,
			OpTag:   balancer.OpUp,
		})
	}

}
*/

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
	dc.discoverImpl()

	for {
		select {
		case <-dc.discoverTicker.C:
			syncService()
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
