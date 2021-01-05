## Braid
**Braid** 提供统一的交互模型，通过注册`模块`（支持自定义），构建属于自己的微服务。

---

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid)](https://goreportcard.com/report/github.com/pojol/braid)
[![drone](http://123.207.198.57:8001/api/badges/pojol/braid-go/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)



### 交互模型
> braid.Mailbox 统一的交互模型

| 共享（多个消息副本 | 竞争（只被消费一次 | 进程内 | 集群内 | 发布 | 订阅 |
| ---- | ---- | ---- | ---- | ---- | ---- |
|Shared | Competition | Proc | Cluster | Pub | Sub |

> `范例` 发布订阅消息

```go
// 订阅一个信道`topic` 这个信道在进程（Proc 内广播（Shared
consumer := braid.Mailbox().Sub(mailbox.Proc, topic).Shared()

// 收取消息
consumer.OnArrived(func(msg mailbox.Message) error {
  return nil
})

// 发送（串行
braid.Mailbox().Pub("topic", message)
// 发送（并行
braid.Mailbox().PubAsync("topic", message)

```

> `范例` 发起一次rpc请求

```go
// ctx 用于分布式追踪，存储调用链路的上下文信息
// target 目标服务节点 例("mail")
// methon 目标节点支持的方法 列("api.mail/send")
// token 调用者的唯一平局（用于链路缓存
// args 输入参数
// reply 回复参数
// opts 调用的额外参数选项
err := braid.GetClient().Invoke(ctx, target, methon, token, args, reply, opts...)
if err != nil {
  // todo ...
}

```

### 微服务
> braid.Module 默认提供的微服务组件

|**Discover**|**Balancer**|**Elector**|**RPC**|**Tracer**|**LinkCache**|
|-|-|-|-|-|-|
|服务发现|负载均衡|选举|RPC|分布式追踪|链路缓存|
|discoverconsul|balancerrandom|electorconsul|grpc-client|[jaegertracer](https://github.com/pojol/braid-go/wiki/Guide-7.-%E4%BD%BF%E7%94%A8Tracer)|[linkerredis](https://github.com/pojol/braid-go/wiki/Guide-4.-%E4%BD%BF%E7%94%A8Link-cahe)|
||[balancerswrr](https://github.com/pojol/braid-go/wiki/Guide-6.-%E8%B4%9F%E8%BD%BD%E5%9D%87%E8%A1%A1)|electork8s|grpc-server|||

### 构建
> 通过注册模块(braid.Module)，构建braid的运行环境。

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



#### Wiki
> Wiki 中提供了一个较为详细的Guide，用于帮助用户更加全面的理解braid的设计理念

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

