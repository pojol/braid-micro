## Braid

---

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid-go)](https://goreportcard.com/report/github.com/pojol/braid-go)
[![CI](https://github.com/pojol/braid-go/actions/workflows/actions.yml/badge.svg?branch=develop)](https://github.com/pojol/braid-go/actions/workflows/actions.yml)
[![Coverage Status](https://coveralls.io/repos/github/pojol/braid-go/badge.svg?branch=develop)](https://coveralls.io/github/pojol/braid-go?branch=develop)
[![](https://img.shields.io/badge/sample-%E6%A0%B7%E4%BE%8B-2ca5e0?style=flat&logo=appveyor)](https://github.com/pojol/braidgo-sample)
[![](https://img.shields.io/badge/doc-%E6%96%87%E6%A1%A3-2ca5e0?style=flat&logo=appveyor)](https://docs.braid-go.fun)
[![](https://img.shields.io/badge/slack-%E4%BA%A4%E6%B5%81-2ca5e0?style=flat&logo=slack)](https://join.slack.com/t/braid-world/shared_invite/zt-mw95pa7m-0Kak8lwE3o4KGMaTuxatJw)


[中文](README_CN.md)

# Intro

> Description of `Service` `Node` `Module` `RPC` `Pub-sub` 

[![image.png](https://i.postimg.cc/nrMjGVGP/image.png)](https://postimg.cc/xND1027v)

---

* **RPC Client/Server** - Used for request / response from `service` to `service` 
* **Pub-sub** - Used to publish & subscribe messages from `module` to `module` 
* **Discover** - Automatic service discovery, and broadcast the node's entry, exit, update and other messages 
* **Balancer** - Client side load balancing which built on service discovery. Provide smooth weighted round-robin balancing by default 
* **Elector** - Select a unique master node for the same name service
* **Tracer** - Distributed tracing system, used to monitor the internal state of the program running in microservices
* **Linkcache** - Link cache used to maintain connection information in distributed systems

### Quick start

```go

b, _ := NewService("braid")

b.RegisterDepend(
	depend.Logger(),
	depend.Redis(redis.WithAddr(mock.RedisAddr)),
	depend.Tracer(
		tracer.WithHTTP(mock.JaegerAddr),
		tracer.WithProbabilistic(1),
	),
	depend.Consul(
		consul.WithAddress([]string{mock.ConsulAddr}),
	),
)

b.RegisterModule(
	module.Pubsub(
		pubsub.WithLookupAddr([]string{mock.NSQLookupdAddr}),
		pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	),
	module.Client(
		client.AppendInterceptors(grpc_prometheus.UnaryClientInterceptor),
	),
	module.Server(
		server.WithListen(":14222"),
		server.AppendInterceptors(grpc_prometheus.UnaryServerInterceptor),
	),
	module.Discover(),
	module.Elector(
		elector.WithLockTick(3*time.Second)),
	module.LinkCache(
		linkcache.WithMode(linkcache.LinkerRedisModeLocal),
	),
)

b.Init()
b.Run()
defer b.Close()

```


### Sample
* RPC - 
	```go
	err := braid.Client().Invoke(
		ctx,
		"target service name (login",
		"methon (/login/guest",
		"token (optional",
		body,
		res,
	)
	```
* Pubsub
	```go
	braid.Pubsub().LocalTopic("topic").Pub(*pubsub.Message)

	lc := braid.Pubsub().LocalTopic("topic").Sub("name")
	lc.Arrived(func(msg *pubsub.Message){ 
		/* todo ... */ 
	})
	defer lc.Close()

	cc := braid.ClusterTopic("topic").Sub("name")
	cc.Arrived(func(msg *pubsub.Message){ 
		/* todo ... */
	})
	defer cc.Close()
	```
* Tracer
	```go
	b.RegisterDepend(
		depend.Tracer(
			tracer.WithHTTP(jaegerAddr),
			tracer.WithProbabilistic(jaegerProbabilistic),
			tracer.WithSpanFactory(
				tracer.TracerFactory{
					Name:    mspan.Mongo,
					Factory: mspan.CreateMongoSpanFactory(),
				},
			),
		),
	)

	span := braid.Tracer().GetSpan(mspan.Mongo)

	span.Begin(ctx)
	defer span.End()

	// todo ...
	span.SetTag("key", val)
	```


#### **Pub-sub** Benchmark
*  ScopeProc

```shell
$ go test -benchmem -run=^$ -bench ^BenchmarkTestProc -cpu 2,4,8
cpu: 2.2 GHz 4 Cores Intel Core i7
goos: darwin
goarch: amd64
pkg: github.com/pojol/braid-go/modules/mailboxnsq
BenchmarkTestProc-2   4340389   302 ns/op   109 B/op   3 allocs/op
BenchmarkTestProc-4   8527536   151 ns/op   122 B/op   3 allocs/op
BenchmarkTestProc-8   7564869   161 ns/op   118 B/op   3 allocs/op
PASS
```

* ScopeCluster

```shell
$ go test -benchmem -run=^$ -bench ^BenchmarkClusterBroadcast -cpu 2,4,8
tencent cloud 4 Cores
goos: linux
goarch: amd64
BenchmarkClusterBroadcast-2   70556   17234 ns/op   540 B/op   16 allocs/op
BenchmarkClusterBroadcast-4   71202   18975 ns/op   676 B/op   20 allocs/op
BenchmarkClusterBroadcast-8   62098   19037 ns/op   662 B/op   20 allocs/op
```