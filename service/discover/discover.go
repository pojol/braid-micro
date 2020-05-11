package discover

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pojol/braid/cache/link"
	"github.com/pojol/braid/consul"
	"github.com/pojol/braid/log"
	"github.com/pojol/braid/service/balancer"
)

type (
	// Sync 服务节点状态管理
	Sync struct {
		ticker *time.Ticker
		cfg    Config

		// service id : service nod
		passingMap map[string]syncNode
	}

	// Config Sync Config
	Config struct {
		Name string

		// 同步节点信息间隔
		Interval int

		ConsulAddress string
	}

	syncNode struct {
		service string
		id      string
		address string
	}
)

var (
	// StatusMgr 状态管理器
	sync *Sync

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

const (
	// DiscoverTag 用于docker发现的tag， 所有希望被discover服务发现的节点，
	// 都应该在Dockerfile中设置 ENV SERVICE_TAGS=braid
	DiscoverTag = "braid"
)

// New 构建指针
func New() *Sync {
	sync = &Sync{}
	return sync
}

// Init init
func (s *Sync) Init(cfg interface{}) error {
	dcfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	s.passingMap = make(map[string]syncNode)
	s.cfg = dcfg

	return nil
}

func (s *Sync) tick() {

	services, err := consul.GetCatalogServices(s.cfg.ConsulAddress, DiscoverTag)
	if err != nil {
		return
	}

	for _, service := range services {
		if service.ServiceName == s.cfg.Name {
			continue
		}

		if _, ok := s.passingMap[service.ServiceID]; !ok { // new nod

			sn := syncNode{
				service: service.ServiceName,
				id:      service.ServiceID,
				address: service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort),
			}

			_, err := link.Get().Num(sn.address)
			if err != nil {
				continue
			}

			s.passingMap[service.ServiceID] = sn

			g, err := balancer.GetGroup(sn.service)
			if err != nil {
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

	for k := range s.passingMap {
		if _, ok := services[k]; !ok { // rmv nod

			g, err := balancer.GetGroup(s.passingMap[k].service)
			if err != nil {
				continue
			}

			g.Rmv(s.passingMap[k].id)

			/*
				err := link.Get().Offline(s.passingMap[k].address)
				if err != nil {
					continue
				}
			*/

			delete(s.passingMap, k)
		}
	}
}

func (s *Sync) runImpl() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("status", "sync", fmt.Errorf("%v", err).Error())
			}
		}()
		// todo ..
		s.tick()
	}

	s.ticker = time.NewTicker(time.Duration(s.cfg.Interval) * time.Millisecond)
	s.tick()

	for {
		select {
		case <-s.ticker.C:
			syncService()
		}
	}
}

// Run 运行管理器
func (s *Sync) Run() {
	go func() {
		s.runImpl()
	}()
}

// Close close
func (s *Sync) Close() {

}
