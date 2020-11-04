## Braid
**Braid** 提供统一的`模块` `服务`交互模型，通过注册`插件`（支持自定义），构建属于自己的微服务。

---

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid)](https://goreportcard.com/report/github.com/pojol/braid)
[![drone](http://123.207.198.57:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)

<img src="https://i.postimg.cc/B6b6CMjM/image.png" width="600">

### 微服务
> braid 提供常用的微服务组件

|  服务  | 简介  |
|  ----  | ----  | 
| **Discover**  | 发现插件，主要提供 `Add` `Rmv` `Update` 等接口 |
| **Elector** | 选举插件，主要提供 `Wait` `Slave` `Master` 等接口 |
| **Rpc** | RPC插件，主要用于发起RPC请求 `Invoke` 和开启RPC Server |
| **Tracer** | 分布式追踪插件，监控分布式系统中的调用细节，目前有`grpc` `echo` `redis` 等追踪 |
| **LinkCache** | 服务访问链路缓存插件，主要用于缓存token（用户唯一凭证）的链路信息 |



### 交互模型
> 使用braid提供的接口或者订阅消息均是多线程安全的

* **同步**
> 在braid中目前只提供一个同步语义的接口,发起一次rpc调用

```go
braid.Invoke()
```

* **异步**
> braid中异步语义的简介，在内部或是将来功能的扩充，都应该优先使用异步语义

| 共享（多个消息副本 | 竞争（只被消费一次 | 进程内 | 集群内 | 发布 | 订阅 |
| ---- | ---- | ---- | ---- | ---- | ---- | ---- |
|Shared | Competition | Proc | Cluster | Pub | Sub |

> `范例` 在集群内`订阅`一个共享型的消息

```go
consumer := braid.Mailbox().ClusterSub(`topic`).AddShared()
consumer.OnArrived(func (msg *mailbox.Message) error {
  return nil
})
```



### 构建
> 通过注册插件，构建braid的运行环境。

```go
b, _ := braid.New(ServiceName)

// 注册插件
b.RegistPlugin(
  braid.Discover(         // Discover 插件
    discoverconsul.Name,  // 插件名（基于consul实现的discover插件
    discoverconsul.WithConsulAddr(consulAddr)), // 插件的可选项
  braid.GRPCClient(grpcclient.Name),
  braid.Elector(
    electorconsul.Name,
    electorconsul.WithConsulAddr(consulAddr),
  ),
  braid.LinkCache(linkerredis.Name),
  braid.JaegerTracing(tracer.WithHTTP(jaegerAddr), tracer.WithProbabilistic(0.01)))

b.Init()  // 初始化注册在braid中的插件
b.Run()   // 运行
defer b.Close() // 释放
```



#### Wiki
https://github.com/pojol/braid/wiki

#### Sample
https://github.com/pojol/braid-sample



#### Web
* 流向图
> 用于监控链路上的连接数以及分布情况

```shell
$ docker pull braidgo/sankey:latest
$ docker run -d -p 8888:8888/tcp braidgo/sankey:latest \
    -consul http://172.17.0.1:8500 \
    -redis redis://172.17.0.1:6379/0
```
<img src="https://i.postimg.cc/sX0xHZmF/image.png" width="600">

