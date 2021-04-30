package balancerswrr

import (
	"strconv"
	"testing"
	"time"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/modules/balancergroupbase"
	"github.com/pojol/braid-go/modules/mailboxnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.M) {

	t.Run()
}

func TestWRR(t *testing.T) {

	serviceName := "TestWRR"

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build(serviceName, log)

	bgb := module.GetBuilder(balancergroupbase.Name)
	bgb.AddOption(balancergroupbase.WithStrategy([]string{Name}))
	b, _ := bgb.Build(serviceName, mb, log)
	bg := b.(balancer.IBalancerGroup)

	b.Init()
	b.Run()
	defer b.Close()

	mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  4,
		Name:    serviceName,
	}))

	mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  2,
		Name:    serviceName,
	}))

	mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
		ID:      "C",
		Address: "C",
		Weight:  1,
		Name:    serviceName,
	}))

	var tests = []struct {
		ID string
	}{
		{"A"}, {"B"}, {"A"}, {"C"}, {"A"}, {"B"}, {"A"},
	}

	time.Sleep(time.Millisecond * 1000)
	for _, v := range tests {
		nod, _ := bg.Pick(Name, serviceName)
		assert.Equal(t, nod.Address, v.ID)
	}

}

func TestWRRDymc(t *testing.T) {

	serviceName := "TestWRRDymc"

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build(serviceName, log)

	bgb := module.GetBuilder(balancergroupbase.Name)
	bgb.AddOption(balancergroupbase.WithStrategy([]string{Name}))
	b, _ := bgb.Build(serviceName, mb, log)
	bg := b.(balancer.IBalancerGroup)

	b.Init()
	b.Run()
	defer b.Close()

	pmap := make(map[string]int)

	mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  1000,
		Name:    serviceName,
	}))

	mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  1000,
		Name:    serviceName,
	}))

	mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
		ID:      "C",
		Address: "C",
		Weight:  1000,
		Name:    serviceName,
	}))

	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 100; i++ {
		nod, _ := bg.Pick(Name, serviceName)
		pmap[nod.ID]++
	}

	mb.Pub(mailbox.Proc, discover.DiscoverUpdateService, mailbox.NewMessage(discover.Node{
		ID:     "A",
		Weight: 500,
	}))
	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 100; i++ {
		nod, _ := bg.Pick(Name, serviceName)
		pmap[nod.ID]++
	}
}

func TestWRROp(t *testing.T) {

	serviceName := "TestWRROp"

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build(serviceName, log)

	bgb := module.GetBuilder(balancergroupbase.Name)
	bgb.AddOption(balancergroupbase.WithStrategy([]string{Name}))
	b, _ := bgb.Build(serviceName, mb, log)
	bg := b.(balancer.IBalancerGroup)

	b.Init()
	b.Run()
	defer b.Close()

	mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
		ID:     "A",
		Name:   serviceName,
		Weight: 4,
	}))

	mb.Pub(mailbox.Proc, discover.DiscoverRmvService, mailbox.NewMessage(discover.Node{
		ID:   "A",
		Name: serviceName,
	}))

	mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
		ID:     "B",
		Name:   serviceName,
		Weight: 2,
	}))

	time.Sleep(time.Millisecond * 500)
	for i := 0; i < 10; i++ {
		nod, err := bg.Pick(Name, serviceName)
		assert.Equal(t, err, nil)
		assert.Equal(t, nod.ID, "B")
	}

}

//20664206	        58.9 ns/op	       0 B/op	       0 allocs/op
func BenchmarkWRR(b *testing.B) {
	serviceName := "BenchmarkWRR"

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build(serviceName, log)

	bgb := module.GetBuilder(balancergroupbase.Name)
	bgb.AddOption(balancergroupbase.WithStrategy([]string{Name}))
	bm, _ := bgb.Build(serviceName, mb, log)
	bg := bm.(balancer.IBalancerGroup)

	bm.Init()
	bm.Run()
	defer bm.Close()

	for i := 0; i < 100; i++ {
		mb.Pub(mailbox.Proc, discover.DiscoverAddService, mailbox.NewMessage(discover.Node{
			ID:     strconv.Itoa(i),
			Name:   serviceName,
			Weight: i,
		}))
	}

	time.Sleep(time.Millisecond * 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bg.Pick(Name, serviceName)
	}
}
