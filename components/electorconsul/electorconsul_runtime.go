// 实现文件 electorconsul 基于 consul 实现的选举
package electorconsul

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid-go/components/depends/bconsul"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
)

const (
	// Name 选举器名称
	Name = "ConsulElection"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("[Elector] convert config error")
)

func (e *consulElection) Init() error {

	sid, err := e.client.CreateSession(e.info.Name + "_lead")
	if err != nil {
		e.log.Warnf("[Elector] %v Dependency check error %v", e.info.Name, err.Error())
		return fmt.Errorf("[Elector] %v Dependency check error %v", e.info.Name, "consul")
	}

	e.sessionID = sid
	e.log.Infof("[Eleactor] create session succ id:%v\n", sid)

	return nil
}

func BuildWithOption(info meta.ServiceInfo, opts ...Option) module.IElector {

	p := Parm{
		LockTick:          time.Second * 2,
		RefushSessionTick: time.Second * 5,
	}

	for _, opt := range opts {
		opt(&p)
	}

	p.Pubsub.GetTopic(info.Name + "." + info.ID + "." + meta.TopicElectionChangeState)

	return &consulElection{
		parm:   p,
		ps:     p.Pubsub,
		client: p.ConsulCli,
		log:    p.Log,
	}

}

type consulElection struct {
	lockTicker   *time.Ticker
	refushTicker *time.Ticker

	client *bconsul.Client

	sessionID string
	locked    bool

	log *blog.Logger
	ps  module.IPubsub

	info meta.ServiceInfo

	parm Parm
}

func (e *consulElection) watch() {
	watchLock := func() {
		defer func() {
			if err := recover(); err != nil {
				e.log.Errf("[Elector] watchLock err %v", err)
			}
		}()

		if !e.locked {
			succ, err := e.client.AcquireLock(e.info.Name, e.sessionID)
			if err != nil {
				e.log.Warnf("[Elector] acquire lock service %s err %v", e.info.Name, err.Error())
			}
			if succ {
				e.locked = true
				e.ps.GetTopic(meta.TopicElectionChangeState).Pub(context.TODO(), meta.EncodeStateChangeMsg(meta.EMaster, e.info.ID))
				e.log.Infof("[Elector] acquire lock service %s, id %s", e.info.Name, e.sessionID)
			} else {
				e.ps.GetTopic(meta.TopicElectionChangeState).Pub(context.TODO(), meta.EncodeStateChangeMsg(meta.ESlave, e.info.ID))
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

		if e.locked {
			refushSession()
		}
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
	e.client.ReleaseLock(e.info.Name, e.sessionID)
	e.client.DeleteSession(e.sessionID)
}
