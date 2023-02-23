package bredis

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

func TestRedis(t *testing.T) {

}
