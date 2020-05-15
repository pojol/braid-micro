package election

import (
	"errors"
	"fmt"
	"time"

	"github.com/pojol/braid/consul"
	"github.com/pojol/braid/log"
)

type (
	// Election 自动选举结构
	Election struct {
		lockTicker   *time.Ticker
		refushTicker *time.Ticker

		sessionID string
		locked    bool

		cfg config
	}

	// IElection 选举器需要提供的接口
	IElection interface {
		IsMaster() bool
	}
)

var (
	e *Election

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

// New 构建新的选举器指针
func New(name string, consulAddress string, opts ...Option) *Election {

	const (
		defaultLockTick   = time.Second * 2000
		defaultRefushTick = time.Second * 5000
	)

	e = &Election{
		cfg: config{
			Name:              name,
			Address:           consulAddress,
			LockTick:          defaultLockTick,
			RefushSessionTick: defaultRefushTick,
		},
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Init 初始化选举器
func (e *Election) Init() error {

	var sid string
	var locked bool

	sid, err := consul.CreateSession(e.cfg.Address, e.cfg.Name+"_lead")
	if err != nil {
		return err
	}

	locked, err = consul.AcquireLock(e.cfg.Address, e.cfg.Name, sid)
	if err != nil {
		return err
	}
	if locked {
		log.SysElection("master")
	} else {
		log.SysElection("slave")
	}

	e.sessionID = sid
	e.locked = locked

	return err
}

// IsMaster 返回是否获取到锁
func (e *Election) IsMaster() bool {
	return e.locked
}

func (e *Election) runImpl() {

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
func (e *Election) Run() {
	go func() {
		e.runImpl()
	}()
}

// Close 释放锁，删除session
func (e *Election) Close() {
	consul.ReleaseLock(e.cfg.Address, e.cfg.Name, e.sessionID)
	consul.DeleteSession(e.cfg.Address, e.sessionID)
}
