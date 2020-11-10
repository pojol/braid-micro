package discoverconsul

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/pojol/braid/3rd/consul"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/modules/linkerredis"
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

func newConsulDiscover() module.Builder {
	return &consulDiscoverBuilder{}
}

func (b *consulDiscoverBuilder) Name() string {
	return Name
}

func (b *consulDiscoverBuilder) Type() string {
	return module.TyDiscover
}

func (b *consulDiscoverBuilder) AddOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *consulDiscoverBuilder) Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (module.IModule, error) {

	p := Parm{
		Tag:                       "braid",
		Name:                      serviceName,
		SyncServicesInterval:      time.Second * 2,
		SyncServiceWeightInterval: time.Second * 10,
		Address:                   "http://127.0.0.1:8500",
	}
	for _, opt := range b.opts {
		opt.(Option)(&p)
	}

	// check address
	_, err := consul.GetCatalogServices(p.Address, p.Tag)
	if err != nil {
		return nil, err
	}

	e := &consulDiscover{
		parm:       p,
		mb:         mb,
		logger:     logger,
		passingMap: make(map[string]*syncNode),
	}

	return e, nil
}

// Discover 发现管理braid相关的节点
type consulDiscover struct {
	discoverTicker   *time.Ticker
	syncWeightTicker *time.Ticker

	// parm
	parm   Parm
	mb     mailbox.IMailbox
	logger logger.ILogger

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
			sn := syncNode{
				service:    service.ServiceName,
				id:         service.ServiceID,
				address:    service.ServiceAddress + ":" + strconv.Itoa(service.ServicePort),
				dyncWeight: 0,
				physWeight: defaultWeight,
			}
			dc.logger.Debugf("new service %s addr %s", service.ServiceName, sn.address)
			dc.passingMap[service.ServiceID] = &sn

			dc.mb.Pub(mailbox.Proc, discover.AddService, mailbox.NewMessage(discover.Node{
				ID:      sn.id,
				Name:    sn.service,
				Address: sn.address,
				Weight:  sn.physWeight,
			}))

		}
	}

	for k := range dc.passingMap {
		if _, ok := services[k]; !ok { // rmv nod
			dc.logger.Debugf("remove service %s id %s", dc.passingMap[k].service, dc.passingMap[k].id)

			dc.mb.Pub(mailbox.Proc, discover.RmvService, mailbox.NewMessage(discover.Node{
				ID:   dc.passingMap[k].id,
				Name: dc.passingMap[k].service,
			}))

			dc.mb.Pub(mailbox.Cluster, linkerredis.LinkerTopicDown, linkcache.EncodeDownMsg(
				dc.passingMap[k].id,
				dc.passingMap[k].service,
				dc.passingMap[k].address,
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

		dc.mb.Pub(mailbox.Proc, discover.UpdateService, mailbox.NewMessage(discover.Node{
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
				dc.logger.Errorf("consul discover syncService err %v", err)
			}
		}()
		// todo ..
		dc.discoverImpl()
	}

	syncWeight := func() {
		defer func() {
			if err := recover(); err != nil {
				dc.logger.Errorf("consul discover syncWeight err %v", err)
			}
		}()

		dc.lock.Lock()
		dc.syncWeight()
		dc.lock.Unlock()
	}

	dc.discoverTicker = time.NewTicker(dc.parm.SyncServicesInterval)
	dc.syncWeightTicker = time.NewTicker(dc.parm.SyncServiceWeightInterval)

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

func (dc *consulDiscover) Init() {
	linknumC, _ := dc.mb.Sub(mailbox.Proc, linkcache.ServiceLinkNum).Shared()
	linknumC.OnArrived(func(msg *mailbox.Message) error {

		lninfo := linkcache.DecodeLinkNumMsg(msg)
		dc.lock.Lock()
		defer dc.lock.Unlock()

		if _, ok := dc.passingMap[lninfo.ID]; ok {
			dc.passingMap[lninfo.ID].linknum = lninfo.Num
		}

		return nil
	})
}

// Discover 运行管理器
func (dc *consulDiscover) Run() {
	go func() {
		dc.runImpl()
	}()
}

// Close close
func (dc *consulDiscover) Close() {

}

func init() {
	module.Register(newConsulDiscover())
}
