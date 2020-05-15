package discover

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pojol/braid/consul"
	"github.com/pojol/braid/log"
	"github.com/pojol/braid/service/balancer"
)

type (
	// Discover 发现管理braid相关的节点
	Discover struct {
		ticker *time.Ticker
		cfg    config

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
func New(name string, consulAddress string, opts ...Option) *Discover {
	const (
		defaultInterval = time.Millisecond * 2000
	)

	dc = &Discover{
		cfg: config{
			Interval: defaultInterval,
		},
	}

	for _, opt := range opts {
		opt(dc)
	}

	return dc
}

// Init init
func (dc *Discover) Init() error {

	dc.passingMap = make(map[string]syncNode)

	return nil
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

			g, err := balancer.GetGroup(sn.service)
			if err != nil {
				log.SysError("discover", "tick add nod", fmt.Errorf("%v", err).Error())
				continue
			}

			g.Add(balancer.Node{
				ID:      sn.id,
				Name:    sn.service,
				Address: sn.address,
				Weight:  1,
			})
		} else { // 看一下是否需要更新权重

		}
	}

	for k := range dc.passingMap {
		if _, ok := services[k]; !ok { // rmv nod

			g, err := balancer.GetGroup(dc.passingMap[k].service)
			if err != nil {
				log.SysError("discover", "tick rmv nod", fmt.Errorf("%v", err).Error())
				continue
			}

			g.Rmv(dc.passingMap[k].id)

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
