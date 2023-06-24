package discoverk8s

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/pojol/braid-go/components/depends/bk8s"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/internal/utils"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
)

type k8sDiscover struct {
	discoverTicker *time.Ticker
	info           meta.ServiceInfo
	parm           Parm

	cli *bk8s.Client
	log *blog.Logger

	pubsub module.IPubsub

	// service id : service nod
	nodemap map[string]*meta.Node

	sync.Mutex
}

func (k *k8sDiscover) Init() error {
	return nil
}

func BuildWithOption(info meta.ServiceInfo, log *blog.Logger, cli *bk8s.Client, ps module.IPubsub, opts ...Option) module.IDiscover {

	p := Parm{
		SyncServicesInterval: time.Second * 2,
		Namespace:            "default",
		Tag:                  "braid",
	}

	for _, opt := range opts {
		opt(&p)
	}

	return &k8sDiscover{
		info:    info,
		cli:     cli,
		log:     log,
		parm:    p,
		pubsub:  ps,
		nodemap: make(map[string]*meta.Node),
	}

}

// 后面使用 k8s 自带的 watch 机制
func (k *k8sDiscover) discoverImpl() {

	k.Lock()
	defer k.Unlock()

	servicesnodes := make(map[string]bool)
	updateflag := false

	services, err := k.cli.ListServices(context.TODO(), k.parm.Namespace)
	if err != nil {
		k.log.Warnf("[braid.discover] err %v", err.Error())
		return
	}

	for _, v := range services {
		if v.Info.Name == "" || len(v.Nodes) == 0 {
			k.log.Warnf("[braid.discover] service %s has no node", v.Info.Name)
			continue
		}

		if !utils.ContainsInSlice(v.Tags, k.parm.Tag) {
			k.log.Debugf("[braid.discover] rule out with service tag %v, self tag %v", v.Tags, k.parm.Tag)
			continue
		}

		if v.Info.Name == k.info.Name {
			k.log.Debugf("[braid.discover] rule out with self")
			continue
		}

		if utils.ContainsInSlice(k.parm.Blacklist, v.Info.Name) {
			k.log.Debugf("[braid.discover] rule out with black list %v", v.Info.Name)
			continue // 排除黑名单节点
		}

		// 添加节点
		for _, nod := range v.Nodes {

			servicesnodes[nod.ID] = true

			if _, ok := k.nodemap[nod.ID]; !ok {

				sn := meta.Node{
					Name:    v.Info.Name,
					ID:      nod.ID,
					Address: nod.Address + ":" + strconv.Itoa(k.parm.getPortWithServiceName(v.Info.Name)),
				}
				k.log.Infof("[braid.discover] new service %s node %s addr %s", v.Info.Name, nod.ID, sn.Address)
				k.nodemap[nod.ID] = &sn

				k.pubsub.GetTopic(meta.TopicDiscoverServiceUpdate).Pub(context.TODO(), meta.EncodeUpdateMsg(
					meta.TopicDiscoverServiceNodeAdd,
					sn,
				))

				updateflag = true

			}

		}
	}

	// 排除节点
	for nodek := range k.nodemap {

		if _, ok := servicesnodes[nodek]; !ok {
			k.log.Infof("[braid.discover] remove service %s node %s", k.nodemap[nodek].Name, k.nodemap[nodek].ID)

			k.pubsub.GetTopic(meta.TopicDiscoverServiceUpdate).Pub(context.TODO(), meta.EncodeUpdateMsg(
				meta.TopicDiscoverServiceNodeRmv,
				*k.nodemap[nodek],
			))

			delete(k.nodemap, nodek)
			updateflag = true
		}

	}

	// 同步节点信息
	if updateflag {

	}

}

func (k *k8sDiscover) discover() {
	syncService := func() {
		defer func() {
			if err := recover(); err != nil {
				k.log.Errf("[braid.discover] syncService err %v", err)
			}
		}()
		// todo ..
		k.discoverImpl()
	}

	k.discoverTicker = time.NewTicker(k.parm.SyncServicesInterval)

	k.discoverImpl()

	for {
		<-k.discoverTicker.C
		syncService()
	}
}

func (k *k8sDiscover) Run() {

	k.log.Debugf("[braid.discover] running ...")

	go func() {
		k.discover()
	}()
}

func (k *k8sDiscover) Close() {

}
