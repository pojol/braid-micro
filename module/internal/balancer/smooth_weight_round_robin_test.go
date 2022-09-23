package balancer

/*
func TestWRR(t *testing.T) {

	serviceName := "TestWRR"

	blog.New(blog.NewWithDefault())
	mb := module.GetBuilder(pubsub.Name).Build(serviceName).(pubsub.IPubsub)

	bgb := module.GetBuilder(Name)
	b := bgb.Build(serviceName,
		moduleparm.WithPubsub(mb))
	bg := b.(balancer.IBalancer)

	bg.Init()
	bg.Run()
	defer bg.Close()

	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventAddService,
			discover.Node{
				ID:      "A",
				Address: "A",
				Name:    serviceName,
				Weight:  4,
			},
		),
	)
	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventAddService,
			discover.Node{
				ID:      "B",
				Address: "B",
				Name:    serviceName,
				Weight:  2,
			},
		),
	)
	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventAddService,
			discover.Node{
				ID:      "C",
				Address: "C",
				Name:    serviceName,
				Weight:  1,
			},
		),
	)

	var tests = []struct {
		ID string
	}{
		{"A"}, {"B"}, {"A"}, {"C"}, {"A"}, {"B"}, {"A"},
	}

	time.Sleep(time.Millisecond * 1000)
	for _, v := range tests {
		nod, _ := bg.Pick(StrategySwrr, serviceName)
		assert.Equal(t, nod.Address, v.ID)
	}

}
*/

/*
func TestWRRDymc(t *testing.T) {

	serviceName := "TestWRRDymc"

	blog.New(blog.NewWithDefault())
	mb := module.GetBuilder(pubsub.Name).Build(serviceName).(pubsub.IPubsub)

	bgb := module.GetBuilder(Name)
	b := bgb.Build(serviceName,
		moduleparm.WithPubsub(mb))
	bg := b.(balancer.IBalancer)

	bg.Init()
	bg.Run()
	defer bg.Close()

	pmap := make(map[string]int)

	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventAddService,
			discover.Node{
				ID:      "A",
				Address: "A",
				Name:    serviceName,
				Weight:  1000,
			},
		),
	)
	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventAddService,
			discover.Node{
				ID:      "B",
				Address: "B",
				Name:    serviceName,
				Weight:  1000,
			},
		),
	)
	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventAddService,
			discover.Node{
				ID:      "C",
				Address: "C",
				Name:    serviceName,
				Weight:  1000,
			},
		),
	)

	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 100; i++ {
		nod, _ := bg.Pick(Name, serviceName)
		pmap[nod.ID]++
	}

	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventUpdateService,
			discover.Node{
				ID:     "A",
				Weight: 500,
			},
		),
	)
	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 100; i++ {
		nod, _ := bg.Pick(StrategySwrr, serviceName)
		pmap[nod.ID]++
	}
}
*/

/*
func TestWRROp(t *testing.T) {

	serviceName := "TestWRROp"

	blog.New(blog.NewWithDefault())
	mb := module.GetBuilder(pubsub.Name).Build(serviceName).(pubsub.IPubsub)

	bgb := module.GetBuilder(Name)
	b := bgb.Build(serviceName,
		moduleparm.WithPubsub(mb))
	bg := b.(balancer.IBalancer)

	bg.Init()
	bg.Run()
	defer bg.Close()

	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventAddService,
			discover.Node{
				ID:     "A",
				Name:   serviceName,
				Weight: 4,
			},
		),
	)

	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventRemoveService,
			discover.Node{
				ID:   "A",
				Name: serviceName,
			},
		),
	)

	mb.GetTopic(discover.ServiceUpdate).Pub(
		discover.EncodeUpdateMsg(discover.EventAddService,
			discover.Node{
				ID:     "B",
				Name:   serviceName,
				Weight: 2,
			},
		),
	)

	time.Sleep(time.Millisecond * 500)
	for i := 0; i < 10; i++ {
		nod, err := bg.Pick(StrategySwrr, serviceName)
		assert.Equal(t, err, nil)
		assert.Equal(t, nod.ID, "B")
	}

}
*/

/*
//20664206	        58.9 ns/op	       0 B/op	       0 allocs/op
func BenchmarkWRR(b *testing.B) {
	serviceName := "BenchmarkWRR"

	blog.New(blog.NewWithDefault())
	mb := module.GetBuilder(pubsub.Name).Build(serviceName).(pubsub.IPubsub)

	bgb := module.GetBuilder(Name)
	bb := bgb.Build(serviceName,
		moduleparm.WithPubsub(mb))
	bg := bb.(balancer.IBalancer)

	bg.Init()
	bg.Run()
	defer bg.Close()

	for i := 0; i < 100; i++ {

		mb.GetTopic(discover.ServiceUpdate).Pub(
			discover.EncodeUpdateMsg(discover.EventAddService,
				discover.Node{
					ID:     strconv.Itoa(i),
					Name:   serviceName,
					Weight: i,
				},
			),
		)

	}

	time.Sleep(time.Millisecond * 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bg.Pick(StrategySwrr, serviceName)
	}
}
*/
