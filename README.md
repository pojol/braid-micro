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
b, _ := NewService(
	"braid",
	uuid.New().String(),
	&components.DefaultDirector{
		Opts: &components.DirectorOpts{
			ClientOpts: []grpcclient.Option{
				grpcclient.AppendInterceptors(grpc_prometheus.UnaryClientInterceptor),
			},
			ServerOpts: []grpcserver.Option{
				grpcserver.WithListen(":14222"),
				grpcserver.AppendInterceptors(grpc_prometheus.UnaryServerInterceptor),
				grpcserver.RegisterHandler(func(srv *grpc.Server) {
					// register grpc handler
				}),
			},
			ElectorOpts: []electorconsul.Option{
				electorconsul.WithLockTick(3 * time.Second),
			},
			LinkcacheOpts: []linkcacheredis.Option{
				linkcacheredis.WithMode(linkcacheredis.LinkerRedisModeLocal),
			},
		},
	},
)

b.Init()
b.Run()
b.Close()
```

* Rpc
```go
err := braid.Send(
	ctx,
	"login", // target service name
	"/user.password", // methon
	"token", // (optional
	body,
	res,
)
```

* Pub
```go
braid.Topic(meta.TopicLinkcacheUnlink).Pub(ctx, &meta.Message(Body : []byte("usertoken")))
```

* Sub
```go
lc, _ := braid.Topic(meta.TopicElectionChangeState).Sub(ctx, "serviceid")
defer lc.Close()

lc.Arrived(func(msg *meta.Message) error { 
	
	scm := meta.DecodeStateChangeMsg(msg)
	if scm.State == elector.EMaster {
		// todo ...
	}

	return nil
})

```

#### **Rpc** Benchmark
```shell

```

#### **Pubsub** Benchmark
```shell
$ go test -benchmem -run=^$ -bench ^BenchmarkPubsub -cpu 2,4,8
cpu: 2.2 GHz 2.5
goos: darwin
goarch: amd64
pkg: github.com/pojol/braid-go/components/pubsubredis
BenchmarkPubsub-2   1959            724452 ns/op            7254 B/op        193 allocs/op
BenchmarkPubsub-4	2506            525298 ns/op            7313 B/op        194 allocs/op
BenchmarkPubsub-8	4233            282358 ns/op            3853 B/op        103 allocs/op
PASS
```