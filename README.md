## Braid
**Braid** 提供统一的`模块` `服务`交互模型，通过注册`插件`（支持自定义），构建属于自己的微服务。

---

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid)](https://goreportcard.com/report/github.com/pojol/braid)
[![drone](http://123.207.198.57:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)

<img src="https://i.postimg.cc/B6b6CMjM/image.png" width="600">


### Plug-ins
**Mailbox** 异步交互组件，支持进程内以及集群中的消息订阅和发送。

**BalancerSWRR** 平滑加权负载均衡器，同时在开启link-cache后,会依据连接数进行权重的调整。

**Discover** 服务发现，提供服务的`进入` `离开` `更新` 消息。

**Elector** 选举，提供节点是否为 `主` 节点的消息。

**Link-cache** 链路缓存，固定基于用户凭证的节点访问链路。用户可以在此基础上做一些业务逻辑优化。

**GRPC** client & server GRPC的封装, 支持连接池，同时可以绑定分布式追踪插件，以及链路缓存插件。

**JaegerTracer** 基于jaeger的分布式追踪服务

---

#### Sample
```golang
b, _ := braid.New(
  NodeName,
  mailboxnsq.WithLookupAddr([]string{nsqLookupAddr}),
  mailboxnsq.WithNsqdAddr([]string{nsqdAddr}))

b.RegistPlugin(
  braid.Discover(
    discoverconsul.Name,
    discoverconsul.WithConsulAddr(consulAddr)),
  braid.Balancer(balancerswrr.Name),
  braid.GRPCClient(grpcclient.Name),
  braid.Elector(
    electorconsul.Name,
    electorconsul.WithConsulAddr(consulAddr),
  ),
  braid.LinkCache(linkerredis.Name),
  braid.JaegerTracing(tracer.WithHTTP(jaegerAddr), tracer.WithProbabilistic(0.01)))

b.Run()
defer b.Close()
```


#### Quick start

> get braid `v1.1.x` Preview version

```bash
go get github.com/pojol/braid@latest
```

> hello,braid

* Architecture based on Docker
  https://github.com/pojol/braid/wiki/quick-start-with-docker

* Architecture based on k8s
  https://github.com/pojol/braid/wiki/quick-start-with-k8s

#### Web
> 流向图，用于监控链路上的连接数以及分布情况
```shell
$ docker pull braidgo/sankey:latest
$ docker run -d -p 8888:8888/tcp braidgo/sankey:latest \
    -consul http://172.17.0.1:8900 \
    -redis redis://172.17.0.1:6379/0
```
<img src="https://i.postimg.cc/sX0xHZmF/image.png" width="600">

#### Wiki
https://github.com/pojol/braid/wiki

#### Sample
https://github.com/pojol/braid-sample