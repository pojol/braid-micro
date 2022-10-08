// 实现文件 electorconsul 基于 consul 实现的选举
package electorconsul

import (
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/pubsub"
)

const (
	// Name 选举器名称
	Name = "ConsulElection"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("[Elector] convert config error")
)

func BuildWithOption(name string, log *blog.Logger, ps pubsub.IPubsub, client *consul.Client, opts ...elector.Option) elector.IElector {

	p := elector.Parm{
		ServiceName:       name,
		LockTick:          time.Second * 2,
		RefushSessionTick: time.Second * 5,
	}
	for _, opt := range opts {
		opt(&p)
	}

	e := &consulElection{
		parm:   p,
		log:    log,
		ps:     ps,
		client: client,
	}

	if client == nil {
		panic(errors.New("[Elector] need depend consul client"))
	}

	e.ps.LocalTopic(elector.TopicChangeState)

	return e
}

func (e *consulElection) Init() error {

	sid, err := e.client.CreateSession(e.parm.ServiceName + "_lead")
	if err != nil {
		return fmt.Errorf("[Elector] %v Dependency check error %v", e.parm.ServiceName, "consul")
	}

	e.sessionID = sid

	return nil
}

type consulElection struct {
	lockTicker   *time.Ticker
	refushTicker *time.Ticker

	client *consul.Client

	sessionID string
	locked    bool

	log *blog.Logger
	ps  pubsub.IPubsub

	parm elector.Parm
}

func (e *consulElection) watch() {
	watchLock := func() {
		defer func() {
			if err := recover(); err != nil {
				e.log.Errf("[Elector] watchLock err %v", err)
			}
		}()

		if !e.locked {
			succ, _ := e.client.AcquireLock(e.parm.ServiceName, e.sessionID)
			if succ {
				e.locked = true
				e.ps.LocalTopic(elector.TopicChangeState).Pub(elector.EncodeStateChangeMsg(elector.EMaster))
				e.log.Infof("[Elector] acquire lock service %s, id %s", e.parm.ServiceName, e.sessionID)
			} else {
				e.ps.LocalTopic(elector.TopicChangeState).Pub(elector.EncodeStateChangeMsg(elector.ESlave))
			}
		}
	}

	// time.Millisecond * 2000
	e.lockTicker = time.NewTicker(e.parm.LockTick)

	for {
		<-e.lockTicker.C
		watchLock()
	}
}

func (e *consulElection) refresh() {

	refushSession := func() {
		defer func() {
			if err := recover(); err != nil {
				e.log.Errf("[Elector] refresh err %v", err)
			}
		}()

		err := e.client.RefreshSession(e.sessionID)
		if err != nil {
			// log
			e.log.Warnf("[Elector] refresh session err %v", err.Error())
		}
	}

	// time.Millisecond * 1000 * 5
	e.refushTicker = time.NewTicker(e.parm.RefushSessionTick)

	for {
		<-e.refushTicker.C
		refushSession()
	}
}

// Run session 状态检查
func (e *consulElection) Run() {
	go func() {
		e.refresh()
	}()

	go func() {
		e.watch()
	}()
}

// Close 释放锁，删除session
func (e *consulElection) Close() {
	e.client.ReleaseLock(e.parm.ServiceName, e.sessionID)
	e.client.DeleteSession(e.sessionID)
}
