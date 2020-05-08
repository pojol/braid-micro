package caller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/pojol/braid/link"
	"github.com/pojol/braid/utility"

	"github.com/pojol/braid/consul"
	"github.com/pojol/braid/log"

	"github.com/opentracing/opentracing-go"

	"github.com/pojol/braid/cache/pool"
	"github.com/pojol/braid/caller/brpc"
	"github.com/pojol/braid/tracer"
	"google.golang.org/grpc"
)

type (
	centerNod struct {
		id      string
		address string
		weight  int
	}

	// Caller 调用器
	Caller struct {
		centerList []centerNod

		cfg Config

		refushTick *time.Ticker

		health bool

		poolMgr sync.Map
		sync.Mutex
	}

	// ICaller caller的抽象接口
	ICaller interface {
		Call(context.Context, string, string, string, []byte) ([]byte, error)
	}

	// Config 调用器配置项
	Config struct {
		ConsulAddress string

		PoolInitNum  int
		PoolCapacity int
		PoolIdle     time.Duration

		Tracing bool
	}
)

var (
	defaultConfig = Config{
		ConsulAddress: "http://127.0.0.1:8500",
		PoolInitNum:   8,
		PoolCapacity:  32,
		PoolIdle:      time.Second * 120,
		Tracing:       false,
	}
	c *Caller

	// ErrServiceNotAvailable 服务不可用，通常是因为没有查询到中心节点(cooridnate)
	ErrServiceNotAvailable = errors.New("caller service not available")

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	// ErrCantFindNode 在注册中心找不到对应的服务节点
	ErrCantFindNode = errors.New("Can't find service node in center")
)

const (
	// CenterTag 用于发现注册中心, 代理注册中心的节点需要在，
	// Dockerfile中设置 ENV SERVICE_TAGS=coordinate
	CenterTag = "coordinate"
)

// New 构建指针
func New() *Caller {
	c = &Caller{}
	return c
}

// Init 通过配置构建调用器
func (c *Caller) Init(cfg interface{}) error {
	callerCfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	if callerCfg.PoolInitNum == 0 {
		callerCfg.PoolInitNum = defaultConfig.PoolInitNum
		callerCfg.PoolCapacity = defaultConfig.PoolCapacity
		callerCfg.PoolIdle = defaultConfig.PoolIdle
	}

	c.cfg = callerCfg

	return nil
}

// Get get caller global pointer
func Get() *Caller {
	return c
}

func (c *Caller) centerExist(id string) bool {

	for _, v := range c.centerList {
		if v.id == id {
			return true
		}
	}

	return false
}

func (c *Caller) getCenter() string {
	clen := len(c.centerList)
	if clen <= 0 {
		log.Fatalf(ErrServiceNotAvailable.Error())
	}

	idx := utility.RandSpace(0, int64(clen-1))
	return c.centerList[idx].address
}

// 监听中心的列表变化，保证本地列表中中心的可用性。
func (c *Caller) runImpl() {

	refush := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("caller", "refush center list", fmt.Errorf("%v", err).Error())
			}
		}()

		services, err := consul.GetCatalogServices(c.cfg.ConsulAddress, CenterTag)
		if err != nil {
			log.SysError("caller", "catalog services err", fmt.Errorf("%v", err).Error())
			return
		}

		c.Lock()
		defer c.Unlock()

		for _, v := range services {
			if c.centerExist(v.ServiceID) == false {
				c.centerList = append(c.centerList, centerNod{
					id:      v.ServiceID,
					address: v.ServiceAddress + ":" + strconv.Itoa(v.ServicePort),
					weight:  0,
				})
			}
		}

		for i := 0; i < len(c.centerList); i++ {

			if _, ok := services[c.centerList[i].id]; !ok {
				c.centerList = append(c.centerList[:i], c.centerList[i+1:]...)
				i--
			}

		}

		if len(c.centerList) > 0 {
			c.health = true
		}

	}

	c.refushTick = time.NewTicker(time.Second * 2)
	refush()

	for {
		select {
		case <-c.refushTick.C:
			refush()
		}
	}
}

// Run run
func (c *Caller) Run() {
	go func() {
		c.runImpl()
	}()
}

// Close 释放调用器
func (c *Caller) Close() {

}

// Call 执行一次rpc调用
func (c *Caller) Call(parentCtx context.Context, nodName string, serviceName string, token string, body []byte) (out []byte, err error) {

	var address string
	var caPool *pool.GRPCPool
	var caConn *pool.ClientConn
	var caCtx context.Context
	var caCancel context.CancelFunc
	var connCtx context.Context
	var connCancel context.CancelFunc
	var method string
	res := new(brpc.RouteRes)

	c.Lock()
	defer c.Unlock()

	if !c.health {
		return out, ErrServiceNotAvailable
	}

	address, err = c.findNode(parentCtx, nodName, serviceName, token)
	if err != nil {
		goto EXT
	}

	caPool, err = c.pool(address)
	if err != nil {
		goto EXT
	}

	connCtx, connCancel = context.WithTimeout(parentCtx, time.Second)
	defer connCancel()
	caConn, err = caPool.Get(connCtx)
	if err != nil {
		goto EXT
	}
	defer caConn.Put()

	caCtx, caCancel = context.WithTimeout(parentCtx, time.Second)
	defer caCancel()

	method = "/brpc.gateway/routing"
	err = caConn.Invoke(caCtx, method, &brpc.RouteReq{
		Nod:     nodName,
		Service: serviceName,
		ReqBody: body,
	}, res)
	if err != nil {
		caConn.Unhealthy()
		goto EXT
	}

EXT:
	if err != nil {
		log.SysError("caller", "do", err.Error())
	}

	return res.ResBody, err
}

// Find 通过查找器获取目标
func (c *Caller) findNode(parentCtx context.Context, nodName string, serviceName string, key string) (string, error) {
	var address string
	var err error

	if key != "" {

		address, err = link.Get().Target(parentCtx, key)
		if err != nil {
			goto EXT
		}

		if address != "" {
			goto EXT
		}

		address, err = c.getNodeWithCoordinate(parentCtx, nodName, serviceName)
		if err != nil {
			goto EXT
		}

		link.Get().Link(parentCtx, key, address)
	} else {
		address, err = c.getNodeWithCoordinate(parentCtx, nodName, serviceName)
		if err != nil {
			goto EXT
		}
	}

EXT:
	if err != nil {
		// log
		log.SysError("caller", "findNode", err.Error())
	}

	return address, err
}

// Pool 获取grpc连接池
func (c *Caller) pool(address string) (p *pool.GRPCPool, err error) {

	factory := func() (*grpc.ClientConn, error) {
		var conn *grpc.ClientConn
		var err error

		if c.cfg.Tracing {
			interceptor := tracer.ClientInterceptor(opentracing.GlobalTracer())
			conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(interceptor))
		} else {
			conn, err = grpc.Dial(address, grpc.WithInsecure())
		}

		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	pi, ok := c.poolMgr.Load(address)
	if !ok {
		p, err = pool.NewGRPCPool(factory, c.cfg.PoolInitNum, c.cfg.PoolCapacity, c.cfg.PoolIdle)
		if err != nil {
			goto EXT
		}

		c.poolMgr.Store(address, p)
		pi = p
	}

	p = pi.(*pool.GRPCPool)

EXT:
	if err != nil {
		log.SysError("caller", "pool", err.Error())
	}

	return p, err
}

func (c *Caller) getNodeWithCoordinate(parentCtx context.Context, nodName string, serviceName string) (string, error) {
	rres := new(brpc.RouteRes)
	var fres struct {
		Address string
	}
	var address string
	var caCtx context.Context
	var caCancel context.CancelFunc
	var dat []byte
	var conn *pool.ClientConn
	method := "/brpc.gateway/routing"

	p, err := c.pool(c.getCenter())
	if err != nil {
		goto EXT
	}
	conn, err = p.Get(context.Background())
	if err != nil {
		goto EXT
	}
	defer conn.Put()

	caCtx, caCancel = context.WithTimeout(parentCtx, time.Second)
	defer caCancel()

	dat, _ = json.Marshal(struct {
		Nod     string
		Service string
	}{nodName, serviceName})

	if conn.Invoke(caCtx, method, &brpc.RouteReq{
		Nod:     "coordinate",
		Service: "find",
		ReqBody: dat,
	}, rres) != nil {
		goto EXT
	}

	json.Unmarshal(rres.ResBody, &fres)
	if fres.Address == "" {
		err = ErrCantFindNode
		goto EXT
	}
	address = fres.Address

EXT:
	if err != nil {
		log.SysError("caller", "getNodeWithCoordinate", err.Error())
	}

	return address, err
}
