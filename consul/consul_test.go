package consul

import (
	"os"
	"testing"

	"github.com/pojol/braid/mock"
)

func TestMain(m *testing.M) {

	mock.Init()

	code := m.Run()
	// 清理测试环境

	os.Exit(code)
}
