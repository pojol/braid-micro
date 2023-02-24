package bconsul

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

func TestBuild(t *testing.T) {

}
