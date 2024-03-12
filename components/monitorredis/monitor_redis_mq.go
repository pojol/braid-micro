package monitorredis

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/pubsubredis"
	"github.com/pojol/braid-go/module"
	"github.com/redis/go-redis/v9"
)

type redisMqMonitor struct {
	parm MqWatchParm

	log    *blog.Logger
	client *redis.Client

	e *echo.Echo
}

type MQInfo struct {
	Name     string `json:"name"`
	Len      int64  `json:"len"`      // 总的消息长度
	NoAckLen int64  `json:"noacklen"` // 未被 ack 的消息长度

	Groups []MQGroupInfo `json:"groups"`
}

type MQGroupInfo struct {
	Name      string `json:"name"`
	Consumers int64  `json:"consumers"`
}

type DrawPos struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

type NodeInfo struct {
	ServiceName string  `json:"service_name"`
	Name        string  `json:"node_name"`
	Num         int64   `json:"player_num"`
	Status      string  `json:"status"`
	Pos         DrawPos `json:"pos"`
}

type SerivceInfo struct {
	Name  string     `json:"service_name"`
	Pos   DrawPos    `json:"pos"`
	Nodes []NodeInfo `json:"nodes"`
}

type NodeRelationInfo struct {
	Org string `json:"org"`
	Tar string `json:"tar"`
}

type RelationInfo struct {
	Sendcnt  int64            `json:"send_cnt"`
	Relation NodeRelationInfo `json:"relation"`
}

type MGServiceInfo struct {
	ServiceLst  []SerivceInfo  `json:"service_lst"`
	RelationLst []RelationInfo `json:"relation_lst"`
}

func BuildWithOption(log *blog.Logger, client *redis.Client, opts ...MqWatchOption) module.IMonitor {

	parm := MqWatchParm{
		Prot: "9999",
	}

	for _, opt := range opts {
		opt(&parm)
	}

	if client == nil {
		panic("monitor redis mq depends redis.client")
	}

	monitor := &redisMqMonitor{
		log:    log,
		client: client,
		parm:   parm,
		e:      echo.New(),
	}

	monitor.e.Use(middleware.CORS())

	return monitor
}

func (rm *redisMqMonitor) Init() error {
	return nil
}

func (rm *redisMqMonitor) mq_watch() ([]MQInfo, error) {

	infoarr := []MQInfo{}

	topics, err := rm.client.SMembers(context.TODO(), pubsubredis.BraidPubsubTopic).Result()
	if err != nil {
		rm.log.Warnf("[braid.monitor] redis mq monitor get topics failed %v", err)
		return infoarr, err
	}

	for _, topic := range topics {
		info, err := rm.client.XInfoStream(context.TODO(), topic).Result()
		if err != nil {
			rm.log.Warnf("[braid.monitor] redis mq monitor get topic %v info failed %v", topic, err)
			continue
		}

		mqinfo := MQInfo{
			Name:   topic,
			Len:    info.Length,
			Groups: []MQGroupInfo{},
		}

		group, _ := rm.client.XInfoGroups(context.TODO(), topic).Result()
		for _, g := range group {

			mqinfo.Groups = append(mqinfo.Groups, MQGroupInfo{
				Name:      g.Name,
				Consumers: g.Consumers,
			})

		}

		infoarr = append(infoarr, mqinfo)
	}

	return infoarr, nil
}

func (rm *redisMqMonitor) services_watch() []MGServiceInfo {
	info := []MGServiceInfo{}
	return info
}

func (rm *redisMqMonitor) Run() {

	rm.e.POST("/mq", func(c echo.Context) error {
		info, _ := rm.mq_watch()
		c.JSON(http.StatusOK, info)
		return nil
	})

	rm.e.POST("/services", func(c echo.Context) error {
		info := rm.services_watch()
		c.JSON(http.StatusOK, info)
		return nil
	})

	go func() {
		rm.e.Start(":" + rm.parm.Prot)
	}()

}

func (rm *redisMqMonitor) Close() {
	// Stop the service gracefully.
	if err := rm.e.Shutdown(context.TODO()); err != nil {
		panic(err)
	}
}

func (rm *redisMqMonitor) Name() string {
	return "redis_mq_monitor"
}
