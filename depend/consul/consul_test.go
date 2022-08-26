package consul

import (
	"os"
	"testing"

	"github.com/pojol/braid-go/mock"
)

func TestMain(m *testing.M) {

	mock.Init()

	code := m.Run()
	// 清理测试环境

	os.Exit(code)
}

func TestServiceRegist(t *testing.T) {

}

func TestServiceDeregist(t *testing.T) {

}

func TestHealth(t *testing.T) {

}

func TestSessionCreate(t *testing.T) {

}

func TestSessionDelete(t *testing.T) {

}

func TestSessionRefush(t *testing.T) {

}

func TestSessionLockAcquire(t *testing.T) {

}

func TestSessionLockRelease(t *testing.T) {

}
