package electorconsul

import (
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid/3rd/consul"
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/elector"
)

const (
	// Name 选举器名称
	Name = "ConsulElection"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

type consulElectionBuilder struct {
	opts []interface{}
}

func newConsulElection() elector.Builder {
	return &consulElectionBuilder{}
}

func (eb *consulElectionBuilder) AddOption(opt interface{}) {
	eb.opts = append(eb.opts, opt)
}

func (eb *consulElectionBuilder) Build(serviceName string) (elector.IElection, error) {

	p := Parm{
		ConsulAddr:        "http://127.0.0.1:8500",
		ServiceName:       serviceName,
		LockTick:          time.Second * 2,
		RefushSessionTick: time.Second * 5,
	}
	for _, opt := range eb.opts {
		opt.(Option)(&p)
	}

	e := &consulElection{
		parm: p,
	}

	sid, err := consul.CreateSession(e.parm.ConsulAddr, e.parm.ServiceName+"_lead")
	if err != nil {
		log.Debugf("create session with consul err %s, addr %s", err.Error(), e.parm.ConsulAddr)
		return nil, err
	}

	locked, err := consul.AcquireLock(e.parm.ConsulAddr, e.parm.ServiceName, sid)
	if err != nil {
		log.Debugf("acquire lock with consul err %s, addr %s", err.Error(), e.parm.ConsulAddr)
		return nil, err
	}
	if locked {
		log.SysElection(e.parm.ServiceName, sid)
	}

	e.sessionID = sid
	e.locked = locked

	return e, nil
}

func (*consulElectionBuilder) Name() string {
	return Name
}

type consulElection struct {
	lockTicker   *time.Ticker
	refushTicker *time.Ticker

	sessionID string
	locked    bool

	parm Parm
}

// IsMaster 返回是否获取到锁
func (e *consulElection) IsMaster() bool {
	return e.locked
}

func (e *consulElection) runImpl() {

	refushSession := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("discover", "refush Session", fmt.Errorf("%v", err).Error())
			}
		}()

		consul.RefushSession(e.parm.ConsulAddr, e.sessionID)
	}

	watchLock := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("discover", "watch lock", fmt.Errorf("%v", err).Error())
			}
		}()

		if !e.locked {
			succ, _ := consul.AcquireLock(e.parm.ConsulAddr, e.parm.ServiceName, e.sessionID)
			if succ {
				e.locked = true
				log.SysElection(e.parm.ServiceName, e.sessionID)
			}
		}
	}

	// time.Millisecond * 1000 * 5
	e.refushTicker = time.NewTicker(e.parm.RefushSessionTick)
	// time.Millisecond * 2000
	e.lockTicker = time.NewTicker(e.parm.LockTick)

	for {
		select {
		case <-e.refushTicker.C:
			refushSession()
		case <-e.lockTicker.C:
			watchLock()
		}
	}
}

// Run session 状态检查
func (e *consulElection) Run() {
	go func() {
		e.runImpl()
	}()
}

// Close 释放锁，删除session
func (e *consulElection) Close() {
	consul.ReleaseLock(e.parm.ConsulAddr, e.parm.ServiceName, e.sessionID)
	consul.DeleteSession(e.parm.ConsulAddr, e.sessionID)
}

func init() {
	elector.Register(newConsulElection())
}
