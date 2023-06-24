## Braid
**Braid** 轻量易读的微服务框架，使用模块化的结构编写，以及提供统一的消息模型。

---

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid-go)](https://goreportcard.com/report/github.com/pojol/braid-go)
[![CI](https://github.com/pojol/braid-go/actions/workflows/actions.yml/badge.svg?branch=develop)](https://github.com/pojol/braid-go/actions/workflows/actions.yml)
[![Coverage Status](https://coveralls.io/repos/github/pojol/braid-go/badge.svg?branch=develop)](https://coveralls.io/github/pojol/braid-go?branch=develop)
[![](https://img.shields.io/badge/sample-%E6%A0%B7%E4%BE%8B-2ca5e0?style=flat&logo=appveyor)](https://github.com/pojol/braidgo-sample)
[![](https://img.shields.io/badge/doc-%E6%96%87%E6%A1%A3-2ca5e0?style=flat&logo=appveyor)](https://docs.braid-go.fun)
[![](https://img.shields.io/badge/slack-%E4%BA%A4%E6%B5%81-2ca5e0?style=flat&logo=slack)](https://join.slack.com/t/braid-world/shared_invite/zt-mw95pa7m-0Kak8lwE3o4KGMaTuxatJw)


# 简介

> 这个图用于描述下面讲述到的几个关键字 `服务` `节点` `模块` `RPC` `Pub-sub`

[![image.png](https://i.postimg.cc/d3GX2X3S/image.png)](https://postimg.cc/CRwTD9J7)

---

> braid 可以通过`模块`的组合，构建出适用于大多数场景的微服务架构，默认提供了如下模块;

* **RPC** - 用于`服务`到`服务`之间的接口调用
* **Pub-sub** - 用于`模块`到`模块`之间的消息发布&接收
* **Discover** - 服务发现，用于感知微服务中各个服务中节点的状态变更（进入，离开，更新权重等，并将变更同步给进程内的其他模块
* **Balancer** - 负载均衡模块，主要用于将 RPC 调用，合理的分配到各个同名服务中
* **Elector** - 选举模块，为注册模块的同名服务，选出一个唯一的主节点
* **Tracer** - 分布式追踪，主要用于监控微服务中程序运行的内部状态
* **Linkcache** - 链路缓存，主要用于维护，传入用户唯一凭证（token，的调用链路，使该 token 的调用 a1->b1->c2 ... 保持不变

### 构建

```go
b, _ := NewService(
		"braid",
		uuid.New().String(),
		&components.DefaultDirector{
			Opts: &components.DirectorOpts{
				// gprc-client
				ClientOpts: []grpcclient.Option{
					grpcclient.AppendInterceptors(grpc_prometheus.UnaryClientInterceptor),
				},
				// grpc-server
				ServerOpts: []grpcserver.Option{
					grpcserver.WithListen(":14222"),
					grpcserver.AppendInterceptors(grpc_prometheus.UnaryServerInterceptor),
					grpcserver.RegisterHandler(func(srv *grpc.Server) {
						// register grpc handler
					}),
				},
				// 选举 elector
				ElectorOpts: []electorconsul.Option{
					electorconsul.WithLockTick(3 * time.Second),
				},
				// 链路缓存
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
braid.Topic(meta.TopicLinkcacheUnlink).Pub(
	ctx, &meta.Message(Body : []byte("usertoken"))
)
```

* Sub
```go
lc, _ := braid.Topic(meta.TopicElectionChangeState).Sub(ctx, "serviceid")
lc.Arrived(func(msg *meta.Message){ 
	
	scm := meta.DecodeStateChangeMsg(msg)
	if scm.State == elector.EMaster {
		// todo ...
	}

})
defer lc.Close()
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
腾讯云 4 Cores
goos: linux
goarch: amd64
BenchmarkClusterBroadcast-2   70556   17234 ns/op   540 B/op   16 allocs/op
BenchmarkClusterBroadcast-4   71202   18975 ns/op   676 B/op   20 allocs/op
BenchmarkClusterBroadcast-8   62098   19037 ns/op   662 B/op   20 allocs/op
```