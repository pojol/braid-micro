// 实现文件 electorconsul 基于 consul 实现的选举
package electorconsul

import (
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/module/elector"
)

const (
	// Name 选举器名称
	Name = "ConsulElection"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

func BuildWithOption(name string, opts ...elector.Option) elector.IElector {

	p := elector.Parm{
		ServiceName:       name,
		LockTick:          time.Second * 2,
		RefushSessionTick: time.Second * 5,
	}
	for _, opt := range opts {
		opt(&p)
	}

	e := &consulElection{
		parm: p,
	}

	e.parm.Ps.GetTopic(elector.TopicChangeState)

	return e
}

func (e *consulElection) Init() error {

	sid, err := e.parm.ConsulClient.CreateSession(e.parm.ServiceName + "_lead")
	if err != nil {
		return fmt.Errorf("%v Dependency check error %v", e.parm.ServiceName, "consul")
	}

	e.sessionID = sid

	return nil
}

type consulElection struct {
	lockTicker   *time.Ticker
	refushTicker *time.Ticker

	sessionID string
	locked    bool

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
			succ, _ := e.parm.ConsulClient.AcquireLock(e.parm.ServiceName, e.sessionID)
			if succ {
				e.locked = true
				e.parm.Ps.GetTopic(elector.TopicChangeState).Pub(elector.EncodeStateChangeMsg(elector.EMaster))
				blog.Debugf("acquire lock service %s, id %s", e.parm.ServiceName, e.sessionID)
			} else {
				e.parm.Ps.GetTopic(elector.TopicChangeState).Pub(elector.EncodeStateChangeMsg(elector.ESlave))
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

		err := e.parm.ConsulClient.RefreshSession(e.sessionID)
		if err != nil {
			// log
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
		e.refush()
	}()

	go func() {
		e.watch()
	}()
}

// Close 释放锁，删除session
func (e *consulElection) Close() {
	e.parm.ConsulClient.ReleaseLock(e.parm.ServiceName, e.sessionID)
	e.parm.ConsulClient.DeleteSession(e.sessionID)
}
