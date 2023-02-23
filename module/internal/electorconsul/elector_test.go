package electorconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/depend/bconsul"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/internal/pubsubnsq"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	mock.Init()
	m.Run()
}

func TestElection(t *testing.T) {

	log := blog.BuildWithOption()

	ps := pubsubnsq.BuildWithOption("", log, pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))

	e := BuildWithOption(
		"TestElection",
		log,
		ps,
		bconsul.BuildWithOption(bconsul.WithAddress([]string{mock.ConsulAddr})),
	)

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}

func TestParm(t *testing.T) {

	log := blog.BuildWithOption()

	ps := pubsubnsq.BuildWithOption("", log, pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))

	el := BuildWithOption(
		"TestElection",
		log,
		ps,
		bconsul.BuildWithOption(bconsul.WithAddress([]string{mock.ConsulAddr})),
		elector.WithLockTick(time.Second),
		elector.WithSessionTick(time.Second),
	)

	consulel := el.(*consulElection)

	assert.Equal(t, consulel.parm.LockTick, time.Second)
	assert.Equal(t, consulel.parm.RefushSessionTick, time.Second)
}
