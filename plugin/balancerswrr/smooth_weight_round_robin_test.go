package balancerswrr

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/plugin/mailboxnsq"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.M) {
	l := log.New(log.Config{
		Mode:   log.DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, log.WithSys(log.Config{
		Mode:   log.DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	t.Run()
}

func TestWRR(t *testing.T) {
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestWRR")
	bb := module.GetBuilder(Name)
	balancer.NewGroup(bb, mb)
	serviceName := "TestWRR"
	addEvent := discover.AddService + "_" + serviceName
	bw := balancer.Get(serviceName)

	mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  4,
		Name:    serviceName,
	}))

	mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  2,
		Name:    serviceName,
	}))

	mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
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
		nod, _ := bw.Pick()
		assert.Equal(t, nod.Address, v.ID)
	}

}

func TestWRRDymc(t *testing.T) {

	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestWRR")
	bb := module.GetBuilder(Name)
	balancer.NewGroup(bb, mb)
	serviceName := "TestWRR"
	addEvent := discover.AddService + "_" + serviceName
	upEvent := discover.UpdateService + "_" + serviceName
	bw := balancer.Get(serviceName)
	pmap := make(map[string]int)

	mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  1000,
		Name:    serviceName,
	}))

	mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  1000,
		Name:    serviceName,
	}))

	mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
		ID:      "C",
		Address: "C",
		Weight:  1000,
		Name:    serviceName,
	}))

	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 100; i++ {
		nod, _ := bw.Pick()
		pmap[nod.ID]++
	}

	fmt.Println("step 1", pmap)

	mb.ProcPub(upEvent, mailbox.NewMessage(discover.Node{
		ID:     "A",
		Weight: 500,
	}))
	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 100; i++ {
		nod, _ := bw.Pick()
		pmap[nod.ID]++
	}

	fmt.Println("step 2", pmap)

}

func TestWRROp(t *testing.T) {

	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestWRR")
	bb := module.GetBuilder(Name)
	balancer.NewGroup(bb, mb)
	serviceName := "TestWRR"
	addEvent := discover.AddService + "_" + serviceName
	rmvEvent := discover.RmvService + "_" + serviceName
	bw := balancer.Get(serviceName)

	mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
		ID:     "A",
		Name:   serviceName,
		Weight: 4,
	}))

	mb.ProcPub(rmvEvent, mailbox.NewMessage(discover.Node{
		ID:   "A",
		Name: serviceName,
	}))

	mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
		ID:     "B",
		Name:   serviceName,
		Weight: 2,
	}))

	bw.Pick()

}

//20664206	        58.9 ns/op	       0 B/op	       0 allocs/op
func BenchmarkWRR(b *testing.B) {
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestWRR")
	bb := module.GetBuilder(Name)
	balancer.NewGroup(bb, mb)
	serviceName := "BenchmarkWRR"
	addEvent := discover.AddService + "_" + serviceName
	bw := balancer.Get(serviceName)

	for i := 0; i < 100; i++ {
		mb.ProcPub(addEvent, mailbox.NewMessage(discover.Node{
			ID:     strconv.Itoa(i),
			Name:   serviceName,
			Weight: i,
		}))
	}

	time.Sleep(time.Millisecond * 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Pick()
	}
}
