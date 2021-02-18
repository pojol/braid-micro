package electorconsul

import (
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid-go/3rd/consul"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
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

func newConsulElection() module.Builder {
	return &consulElectionBuilder{}
}

func (eb *consulElectionBuilder) AddOption(opt interface{}) {
	eb.opts = append(eb.opts, opt)
}

func (eb *consulElectionBuilder) Build(serviceName string, mb mailbox.IMailbox, logger logger.ILogger) (module.IModule, error) {

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
		parm:   p,
		mb:     mb,
		logger: logger,
	}

	return e, nil
}

func (e *consulElection) Init() error {
	sid, err := consul.CreateSession(e.parm.ConsulAddr, e.parm.ServiceName+"_lead")
	if err != nil {
		return fmt.Errorf("Dependency check error %v [%v]", "consul", e.parm.ConsulAddr)
	}

	e.sessionID = sid

	return nil
}

func (*consulElectionBuilder) Name() string {
	return Name
}

func (*consulElectionBuilder) Type() string {
	return module.TyElector
}

type consulElection struct {
	lockTicker   *time.Ticker
	refushTicker *time.Ticker

	sessionID string
	locked    bool

	logger logger.ILogger

	mb   mailbox.IMailbox
	parm Parm
}

func (e *consulElection) watch() {
	watchLock := func() {
		defer func() {
			if err := recover(); err != nil {
				e.logger.Errorf("discover watchLock err %v", err)
			}
		}()

		if !e.locked {
			succ, _ := consul.AcquireLock(e.parm.ConsulAddr, e.parm.ServiceName, e.sessionID)
			if succ {
				e.locked = true
				e.mb.Pub(mailbox.Proc, elector.StateChange, elector.EncodeStateChangeMsg(elector.EMaster))
				e.logger.Debugf("acquire lock service %s, id %s", e.parm.ServiceName, e.sessionID)
			} else {
				e.mb.Pub(mailbox.Proc, elector.StateChange, elector.EncodeStateChangeMsg(elector.ESlave))
			}
		}
	}

	watchLock()

	// time.Millisecond * 2000
	e.lockTicker = time.NewTicker(e.parm.LockTick)

	for {
		select {
		case <-e.lockTicker.C:
			watchLock()
		}
	}
}

func (e *consulElection) refush() {

	refushSession := func() {
		defer func() {
			if err := recover(); err != nil {
				e.logger.Errorf("discover refush err %v", err)
			}
		}()

		consul.RefushSession(e.parm.ConsulAddr, e.sessionID)
	}

	// time.Millisecond * 1000 * 5
	e.refushTicker = time.NewTicker(e.parm.RefushSessionTick)

	for {
		select {
		case <-e.refushTicker.C:
			refushSession()
		}
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

func init() {
	module.Register(newConsulElection())
}
