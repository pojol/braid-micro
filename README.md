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

### Modules

|**Discovery**|**Balancing**|**Elector**|**RPC**|**Pub-sub**|**Tracer**|**LinkCache**|
|-|-|-|-|-|-|-|
|discoverconsul|balancerrandom|electorconsul|grpc-client|mailbox|jaegertracer|linkerredis
||balancerswrr|electork8s|grpc-server|||

### Quick start

```go

s := braid.NewService("gate")   // create a new node in the service gate

// register module in node
s.Register(
    braid.Module(braid.LoggerZap),
    braid.Module(braid.PubsubNsq,
        pubsubnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}),
        pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
    ),
    braid.Module(
        braid.DiscoverConsul,    // discover module
        discoverconsul.WithConsulAddr(consulAddr)
    ),
)

s.Init()
s.Run()

defer s.Close()

```



#### Web
* Shankey chart
> Monitor the connection in the cluster

```shell
$ docker pull braidgo/sankey:latest
$ docker run -d -p 8888:8888/tcp braidgo/sankey:latest \
    -consul http://172.17.0.1:8500 \
    -redis redis://172.17.0.1:6379/0
```
<img src="https://i.postimg.cc/sX0xHZmF/image.png" width="600">



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