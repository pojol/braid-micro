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
		lockTicker    *time.Ticker
		sessionTicker *time.Ticker

		sessionID string
		locked    bool

		// consul address
		address string
		// nod name (nod name)
		nodName string
	}

	Config struct {
		Address string
		Name    string
	}
)

var (
	e *Election

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")
)

func New() *Election {
	e = &Election{}
	return e
}

// Init 初始化选举器
func (e *Election) Init(cfg interface{}) error {

	elCfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

	var sid string
	var locked bool

	sid, err := consul.CreateSession(elCfg.Address, elCfg.Name+"_lead")
	if err != nil {
		return err
	}

	locked, err = consul.AcquireLock(elCfg.Address, elCfg.Name, sid)
	if err != nil {
		return err
	}
	if locked {
		fmt.Println("master")
	} else {
		fmt.Println("slave")
	}

	e.sessionID = sid
	e.locked = locked
	e.nodName = elCfg.Name
	e.address = elCfg.Address

	return err
}

// IsLocked 返回是否获取到锁
func (e *Election) IsLocked() bool {
	return e.locked
}

func (e *Election) runImpl() {

	refushSession := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("discover", "refush Session", fmt.Errorf("%v", err).Error())
			}
		}()

		consul.RefushSession(e.address, e.sessionID)
	}

	watchLock := func() {
		defer func() {
			if err := recover(); err != nil {
				log.SysError("discover", "watch lock", fmt.Errorf("%v", err).Error())
			}
		}()

		if !e.locked {
			succ, _ := consul.AcquireLock(e.address, e.nodName, e.sessionID)
			if succ {
				e.locked = true
				fmt.Println("master now")
			}
		}

	}

	e.sessionTicker = time.NewTicker(time.Millisecond * 1000 * 5)
	e.lockTicker = time.NewTicker(time.Millisecond * 2000)

	for {
		select {
		case <-e.sessionTicker.C:
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
	consul.ReleaseLock(e.address, e.nodName, e.sessionID)
	consul.DeleteSession(e.address, e.sessionID)
}
