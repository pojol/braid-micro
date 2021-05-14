## Braid
**Braid** 轻量易读的微服务框架，使用模块化的结构编写，以及提供统一的消息模型。

---

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid-go)](https://goreportcard.com/report/github.com/pojol/braid-go)
![workflow](https://github.com/pojol/braid-go/actions/workflows/actions.yml/badge.svg)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)
[![](https://img.shields.io/badge/sample-%E6%A0%B7%E4%BE%8B-2ca5e0?style=flat&logo=appveyor)](https://github.com/pojol/braidgo-sample)
[![](https://img.shields.io/badge/doc-%E6%96%87%E6%A1%A3-2ca5e0?style=flat&logo=appveyor)](https://docs.braid-go.fun)
[![](https://img.shields.io/badge/slack-%E4%BA%A4%E6%B5%81-2ca5e0?style=flat&logo=slack)](https://join.slack.com/t/braid-world/shared_invite/zt-mw95pa7m-0Kak8lwE3o4KGMaTuxatJw)


### 消息模型

#### API 消息
> API消息主要用于**服务**`->`**服务**之间

* ctx 用于分布式追踪，存储调用链路的上下文信息
* target 目标服务节点 例 "mail", braid 会依据服务发现和负载均衡信息，自动将消息发送到合适的节点
* methon 目标节点支持的方法
* token 如果为`""`则随机指向一个目标服务，如果传入用户唯一凭据则可以使用平滑加权负载均衡，以及链路缓存
* args 输入参数
* reply 回复参数
* opts... gpc调用的额外参数选项

```go
client.Invoke(ctx, target, methon, token, args, reply, opts...)
```

#### Pub-sub 消息
> Pub-sub消息主要用于**模块**`->`**模块**之间

* 作用域
    * mailbox.ScopeProc 消息作用于`自身进程`中的模块
    * mailbox.ScopeCluster 消息将作用于`整个集群`中的模块
* Topic
    > 某个消息的集合，当调用 pub 发布消息时，消息会被投递到加入到这个 topic 的所有 channel 中
    * 单一接收 （一个topic `+` channel x 1 : consumer x 1
    * 广播逻辑 （一个topic `+` channel x N : consumer x N
    * 竞争接收 （一个topic `+` channel x 1 : consumer x N
* Channel
    > topic 中的管道，每一个管道都会接收到来自topic 的消息，并且每个管道都可以拥有多个`消费者`

####  示例
* `多个`消费者`单个` Channel

```go

topic := "test.procNotify"

mailbox.RegistTopic(topic, mailbox.ScopeProc)

mailbox.GetTopic(topic).Sub("channel_1").Arrived(func(msg *mailbox.Message) {
    fmt.Println("consumer a receive", string(msg.Body))
})
mailbox.GetTopic(topic).Sub("channel_1").Arrived(func(msg *mailbox.Message) {
    fmt.Println("consumer b receive", string(msg.Body))
})

for i := 0; i < 5; i++ {
    mailbox.GetTopic(topic).Pub(&mailbox.Message{Body: []byte(strconv.Itoa(i))})
}
```
```shell
# 两个consumer从一个管道中竞争获取消息
consumer a receive 0
consumer b receive 1
consumer a receive 2
consumer b receive 3
consumer a receive 4
```



### 模块
> 默认提供的微服务组件，[**文档地址**](https://docs.braid-go.fun/)

|**Discovery**|**Balancing**|**Elector**|**RPC**|**Pub-sub**|**Tracer**|**LinkCache**|
|-|-|-|-|-|-|-|
|服务发现|负载均衡|选举|RPC|发布-订阅|分布式追踪|链路缓存|
|discoverconsul|balancerrandom|electorconsul|grpc-client|mailbox|jaegertracer|linkerredis
||balancerswrr|electork8s|grpc-server|||

### 构建
> 构建braid的运行环境。

```go
b, _ := braid.New(ServiceName)

// 将模块注册到braid
b.RegistModule(
  braid.Discover(         // Discover 模块
    discoverconsul.Name,  // 模块名（基于consul实现的discover模块，通过模块名可以获取到模块的构建器
    discoverconsul.WithConsulAddr(consulAddr)), // 模块的可选项
  braid.Client(grpcclient.Name),
  braid.Elector(
    electorconsul.Name,
    electorconsul.WithConsulAddr(consulAddr),
  ),
  braid.LinkCache(linkerredis.Name),
  braid.Tracing(
    jaegertracing.Name,
    jaegertracing.WithHTTP(jaegerAddr), 
    jaegertracing.WithProbabilistic(0.01)))

b.Init()  // 初始化注册在braid中的模块
b.Run()   // 运行
defer b.Close() // 释放
```





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

