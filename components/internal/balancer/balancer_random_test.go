package balancer

/*
func TestRandomBalancer(t *testing.T) {

	serviceName := "TestRandomBalancer"
	mock.Init()

	blog.BuildWithNormal()
	ps := pubsub.BuildWithOption(
		serviceName,
		pubsub.WithLookupAddr([]string{mock.NSQLookupdAddr}),
		pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)

	bg := BuildWithOption(serviceName, ps)

	bg.Init()
	bg.Run()
	defer bg.Close()

	var atick, btick, ctick uint64
	_, err := bg.Pick(StrategyRandom, serviceName)
	assert.NotEqual(t, err, nil)

	ps.GetTopic(service.TopicDiscoverServiceUpdate).Pub(service.DiscoverEncodeUpdateMsg(
		discover.EventAddService,
		meta.Node{
			ID:      "A",
			Address: "A",
			Weight:  4,
			Name:    serviceName,
		},
	))
	ps.GetTopic(service.TopicDiscoverServiceUpdate).Pub(service.DiscoverEncodeUpdateMsg(
		discover.EventAddService,
		meta.Node{
			ID:      "B",
			Address: "B",
			Weight:  2,
			Name:    serviceName,
		},
	))
	ps.GetTopic(service.TopicDiscoverServiceUpdate).Pub(service.DiscoverEncodeUpdateMsg(
		discover.EventAddService,
		meta.Node{
			ID:      "C",
			Address: "C",
			Weight:  1,
			Name:    serviceName,
		},
	))

	time.Sleep(time.Millisecond * 100)
	fmt.Println("begin pick")
	for i := 0; i < 30000; i++ {
		nod, err := bg.Pick(StrategyRandom, serviceName)
		assert.Equal(t, err, nil)

		if nod.ID == "A" {
			atomic.AddUint64(&atick, 1)
		} else if nod.ID == "B" {
			atomic.AddUint64(&btick, 1)
		} else if nod.ID == "C" {
			atomic.AddUint64(&ctick, 1)
		}
	}

	fmt.Println(atomic.LoadUint64(&atick), atomic.LoadUint64(&btick), atomic.LoadUint64(&ctick))
	assert.Equal(t, true, (atomic.LoadUint64(&atick) >= 9000 && atomic.LoadUint64(&atick) <= 11000))
	assert.Equal(t, true, (atomic.LoadUint64(&btick) >= 9000 && atomic.LoadUint64(&btick) <= 11000))
	assert.Equal(t, true, (atomic.LoadUint64(&ctick) >= 9000 && atomic.LoadUint64(&ctick) <= 11000))

	ps.GetTopic(service.TopicDiscoverServiceUpdate).Pub(service.DiscoverEncodeUpdateMsg(
		discover.EventRemoveService,
		meta.Node{
			ID:      "C",
			Address: "C",
			Name:    serviceName,
		},
	))

	time.Sleep(time.Millisecond * 100)
	atomic.SwapUint64(&atick, 0)
	atomic.SwapUint64(&btick, 0)
	atomic.SwapUint64(&ctick, 0)

	for i := 0; i < 20000; i++ {
		nod, _ := bg.Pick(StrategyRandom, serviceName)
		if nod.ID == "A" {
			atomic.AddUint64(&atick, 1)
		} else if nod.ID == "B" {
			atomic.AddUint64(&btick, 1)
		} else if nod.ID == "C" {
			atomic.AddUint64(&ctick, 1)
		}
	}
	assert.Equal(t, true, (atomic.LoadUint64(&atick) >= 9000 && atomic.LoadUint64(&atick) <= 11000))
	assert.Equal(t, true, (atomic.LoadUint64(&btick) >= 9000 && atomic.LoadUint64(&btick) <= 11000))
	assert.Equal(t, true, (atomic.LoadUint64(&ctick) == 0))
}

*/
