package discover

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pojol/braid/balancer"
	"github.com/pojol/braid/consul"
	"github.com/pojol/braid/link"
	"github.com/pojol/braid/log"
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
	// DiscoverTag 适用于docker发现， 在Dockerfile 中设置ENV SERVICE_TAGS=braid
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
		if _, ok := s.passingMap[service.ServiceID]; !ok { // new nod

			sn := syncNode{
				service: service.ServiceName,
				id:      service.ServiceID,
				address: service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort),
			}

			num, err := link.Get().Num(sn.address)
			if err != nil {
				continue
			}

			s.passingMap[service.ServiceID] = sn

			balancer.GetSelector(sn.service).Add(balancer.Node{
				ID:      sn.id,
				Address: sn.address,
				Weight:  1, // 暂时不在提供weight支持
				Tick:    num,
			})
		}
	}

	for k := range s.passingMap {
		if _, ok := services[k]; !ok { // rmv nod

			// 从平衡器中删除节点
			balancer.GetSelector(s.passingMap[k].service).Rmv(s.passingMap[k].id)

			err := link.Get().Offline(s.passingMap[k].address)
			if err != nil {
				continue
			}

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
