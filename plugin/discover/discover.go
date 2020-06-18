package discover

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pojol/braid/3rd/consul"
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/internal/balancer"
)

type (
	// IDiscover 服务发现
	IDiscover interface {
		Run()
		Close()
	}

	// Discover 发现管理braid相关的节点
	Discover struct {
		ticker *time.Ticker
		cfg    config
		bg     *balancer.Group

		// service id : service nod
		passingMap map[string]syncNode
	}

	syncNode struct {
		service string
		id      string
		address string
	}
)

var (
	dc *Discover

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

const (
	// DiscoverTag 用于docker发现的tag， 所有希望被discover服务发现的节点，
	// 都应该在Dockerfile中设置 ENV SERVICE_TAGS=braid
	DiscoverTag = "braid"
)

// New 构建指针
func New(name string, consulAddress string, bg *balancer.Group, opts ...Option) IDiscover {
	const (
		defaultInterval = time.Millisecond * 2000
	)

	dc = &Discover{
		cfg: config{
			Interval:      defaultInterval,
			ConsulAddress: consulAddress,
		},
		bg: bg,
	}

	for _, opt := range opts {
		opt(dc)
	}

	dc.passingMap = make(map[string]syncNode)

	return dc
}

func (dc *Discover) tick() {

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

			/*
				_, err := link.Get().Num(sn.address)
				if err != nil {
					continue
				}
			*/

			dc.passingMap[service.ServiceID] = sn
			dc.bg.Get(sn.service).Update(balancer.Node{
				ID:      sn.id,
				Name:    sn.service,
				Address: sn.address,
				Weight:  1,
				OpTag:   balancer.OpAdd,
			})

		} else { // 看一下是否需要更新权重

		}
	}

	for k := range dc.passingMap {
		if _, ok := services[k]; !ok { // rmv nod

			dc.bg.Get(dc.passingMap[k].service).Update(balancer.Node{
				ID:    dc.passingMap[k].id,
				Name:  dc.passingMap[k].service,
				OpTag: balancer.OpRmv,
			})

			/*
				err := link.Get().Offline(s.passingMap[k].address)
				if err != nil {
					continue
				}
			*/

			delete(dc.passingMap, k)
		}
	}
}

func (dc *Discover) runImpl() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("status", "sync", fmt.Errorf("%v", err).Error())
			}
		}()
		// todo ..
		dc.tick()
	}

	dc.ticker = time.NewTicker(dc.cfg.Interval)
	dc.tick()

	for {
		select {
		case <-dc.ticker.C:
			syncService()
		}
	}
}

// Run 运行管理器
func (dc *Discover) Run() {
	go func() {
		dc.runImpl()
	}()
}

// Close close
func (dc *Discover) Close() {

}
