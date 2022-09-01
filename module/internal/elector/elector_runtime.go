// 实现文件 electorconsul 基于 consul 实现的选举
package elector

import (
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/depend/pubsub"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/service"
)

const (
	// Name 选举器名称
	Name = "ConsulElection"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

func Build(name string, ps pubsub.IPubsub, opts ...elector.Option) elector.IElector {

	p := elector.Parm{
		ConsulAddr:        "http://127.0.0.1:8500",
		ServiceName:       name,
		LockTick:          time.Second * 2,
		RefushSessionTick: time.Second * 5,
	}
	for _, opt := range opts {
		opt(&p)
	}

	e := &consulElection{
		parm: p,
		ps:   ps,
	}

	e.ps.GetTopic(service.TopicElectorChangeState)

	return e
}

func (e *consulElection) Init() error {
	sid, err := consul.CreateSession(e.parm.ConsulAddr, e.parm.ServiceName+"_lead")
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v [%v]", e.parm.ServiceName, "consul", e.parm.ConsulAddr)
	}

	e.sessionID = sid

	return nil
}

type consulElection struct {
	lockTicker   *time.Ticker
	refushTicker *time.Ticker

	sessionID string
	locked    bool

	ps   pubsub.IPubsub
	parm elector.Parm
}

func (e *consulElection) watch() {
	watchLock := func() {
		defer func() {
			if err := recover(); err != nil {
				blog.Errf("discover watchLock err %v", err)
			}
		}()

		if !e.locked {
			succ, _ := consul.AcquireLock(e.parm.ConsulAddr, e.parm.ServiceName, e.sessionID)
			if succ {
				e.locked = true
				e.ps.GetTopic(service.TopicElectorChangeState).Pub(service.ElectorEncodeStateChangeMsg(elector.EMaster))
				blog.Debugf("acquire lock service %s, id %s", e.parm.ServiceName, e.sessionID)
			} else {
				e.ps.GetTopic(service.TopicElectorChangeState).Pub(service.ElectorEncodeStateChangeMsg(elector.ESlave))
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

func (e *consulElection) refush() {

	refushSession := func() {
		defer func() {
			if err := recover(); err != nil {
				blog.Errf("discover refush err %v", err)
			}
		}()

		consul.RefushSession(e.parm.ConsulAddr, e.sessionID)
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
		e.refush()
	}()

	go func() {
		e.watch()
	}()
}

// Close 释放锁，删除session
func (e *consulElection) Close() {
	consul.ReleaseLock(e.parm.ConsulAddr, e.parm.ServiceName, e.sessionID)
	consul.DeleteSession(e.parm.ConsulAddr, e.sessionID)
}
