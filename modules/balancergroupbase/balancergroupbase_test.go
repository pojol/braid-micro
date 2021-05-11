package balancergroupbase

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/modules/balancerrandom"
	"github.com/pojol/braid-go/modules/mailboxnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestParm(t *testing.T) {
	serviceName := "TestParm"

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build(serviceName, log)

	bgb := module.GetBuilder(Name)
	bgb.AddOption(WithStrategy([]string{balancerrandom.Name}))
	b, _ := bgb.Build(serviceName, mb, log)
	bg := b.(*baseBalancerGroup)

	assert.Equal(t, bg.parm.strategies, []string{balancerrandom.Name})

	b.Init()
	b.Run()
	defer b.Close()

	mb.GetTopic(discover.AddService).Pub(mailbox.NewMessage(discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  4,
		Name:    serviceName,
	}))

	mb.GetTopic(discover.AddService).Pub(mailbox.NewMessage(discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  2,
		Name:    serviceName,
	}))

	time.Sleep(time.Millisecond * 100)
	mb.GetTopic(discover.UpdateService).Pub(mailbox.NewMessage(discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  3,
		Name:    serviceName,
	}))

	mb.GetTopic(discover.RemoveService).Pub(mailbox.NewMessage(discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  2,
		Name:    serviceName,
	}))

	time.Sleep(time.Millisecond * 500)
	for i := 0; i < 10; i++ {
		nod, err := bg.Pick(balancerrandom.Name, serviceName)
		if err != nil {
			t.FailNow()
		}
		assert.Equal(t, nod.ID, "A")
	}
}
