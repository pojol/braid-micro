package consulelection

import (
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid/3rd/consul"
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/election"
)

const (
	// ElectionName 选举器名称
	ElectionName = "ConsulElection"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

type consulElectionBuilder struct{}

func newConsulElection() election.Builder {
	return &consulElectionBuilder{}
}

func (*consulElectionBuilder) Build(cfg interface{}) election.IElection {
	cecfg, ok := cfg.(Cfg)
	if !ok {
		return nil
	}

	e := &consulElection{
		cfg: cecfg,
	}

	// verfiy

	sid, err := consul.CreateSession(e.cfg.Address, e.cfg.Name+"_lead")
	if err != nil {
		return nil
	}

	locked, err := consul.AcquireLock(e.cfg.Address, e.cfg.Name, sid)
	if err != nil {
		return nil
	}
	if locked {
		log.SysElection("master")
	} else {
		log.SysElection("slave")
	}

	e.sessionID = sid
	e.locked = locked

	return e
}

func (*consulElectionBuilder) Name() string {
	return ElectionName
}

// Cfg 选举器配置项
type Cfg struct {
	Address           string
	Name              string
	LockTick          time.Duration
	RefushSessionTick time.Duration
}

type consulElection struct {
	lockTicker   *time.Ticker
	refushTicker *time.Ticker

	sessionID string
	locked    bool

	cfg Cfg
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

		consul.RefushSession(e.cfg.Address, e.sessionID)
	}

	watchLock := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("discover", "watch lock", fmt.Errorf("%v", err).Error())
			}
		}()

		if !e.locked {
			succ, _ := consul.AcquireLock(e.cfg.Address, e.cfg.Name, e.sessionID)
			if succ {
				e.locked = true
				log.SysElection("master")
			}
		}
	}

	// time.Millisecond * 1000 * 5
	e.refushTicker = time.NewTicker(e.cfg.RefushSessionTick)
	// time.Millisecond * 2000
	e.lockTicker = time.NewTicker(e.cfg.LockTick)

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
	consul.ReleaseLock(e.cfg.Address, e.cfg.Name, e.sessionID)
	consul.DeleteSession(e.cfg.Address, e.sessionID)
}

func init() {
	election.Register(newConsulElection())
}
