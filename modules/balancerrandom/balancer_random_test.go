package balancerrandom

import (
	"testing"
	"time"

	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/modules/balancergroupbase"
	"github.com/pojol/braid/modules/mailboxnsq"
	"github.com/pojol/braid/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.M) {
	t.Run()
}

func TestRandomBalancer(t *testing.T) {

	serviceName := "TestRandomBalancer"

	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build(serviceName)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()

	bgb := module.GetBuilder(balancergroupbase.Name)
	bgb.AddOption(balancergroupbase.WithStrategy([]string{Name}))
	b, _ := bgb.Build(serviceName, mb, log)
	bg := b.(balancer.IBalancerGroup)

	b.Init()
	b.Run()
	defer b.Close()

	var atick, btick, ctick uint64
	_, err := bg.Pick(Name, serviceName)
	assert.NotEqual(t, err, nil)

	mb.Pub(mailbox.Proc, discover.AddService, mailbox.NewMessage(discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  4,
		Name:    serviceName,
	}))

	mb.Pub(mailbox.Proc, discover.AddService, mailbox.NewMessage(discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  2,
		Name:    serviceName,
	}))

	mb.Pub(mailbox.Proc, discover.AddService, mailbox.NewMessage(discover.Node{
		ID:      "C",
		Address: "C",
		Weight:  1,
		Name:    serviceName,
	}))

	time.Sleep(time.Millisecond * 100)
	for i := 0; i < 30000; i++ {
		nod, _ := bg.Pick(Name, serviceName)
		if nod.ID == "A" {
			atick++
		} else if nod.ID == "B" {
			btick++
		} else if nod.ID == "C" {
			ctick++
		}
	}

	assert.Equal(t, true, (atick >= 9000 && atick <= 11000))
	assert.Equal(t, true, (btick >= 9000 && btick <= 11000))
	assert.Equal(t, true, (ctick >= 9000 && ctick <= 11000))

	mb.Pub(mailbox.Proc, discover.RmvService, mailbox.NewMessage(discover.Node{
		ID:      "C",
		Address: "C",
		Name:    serviceName,
	}))

	time.Sleep(time.Millisecond * 100)
	atick = 0
	btick = 0
	ctick = 0

	for i := 0; i < 20000; i++ {
		nod, _ := bg.Pick(Name, serviceName)
		if nod.ID == "A" {
			atick++
		} else if nod.ID == "B" {
			btick++
		} else if nod.ID == "C" {
			ctick++
		}
	}
	assert.Equal(t, true, (atick >= 9000 && atick <= 11000))
	assert.Equal(t, true, (btick >= 9000 && btick <= 11000))
	assert.Equal(t, true, (ctick == 0))
}
