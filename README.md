## braid
轻量的微服务框架，提供常用的组件，使用braid将使我们专注在实现上，而不需要关心主从，添加删除服务，调度，负载均衡等微服务逻辑。

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/41686ce5edf844fc8b81cffd13cc0550)](https://app.codacy.com/manual/pojol/braid?utm_source=github.com&utm_medium=referral&utm_content=pojol/braid&utm_campaign=Badge_Grade_Dashboard)
[![drone](http://47.96.147.176:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)

> `注:`当前v1.1.x版本为`原型`版本 

#### 组件
* **选举** (election
```go
// 通过 "节点名" "Consul" "LockTick 选举竞争频率" 构建election
elec := election.New("NodeName", ConsulAddress, WithLockTick(1000))

elec.Run()
defer elec.Close()

// 获取当前节点是否为Master节点（Master节点只会存在一个，且当Master节点下线后会从其他同名节点中选举出新的Master.
elec.IsMaster()
```

* **负载均衡** (balancer
> 这是一个内部支持模块，在使用RPC请求时，调用会通过负载均衡来选择一个合适的节点发送。
> 默认使用的是`平滑加权轮询`算法

* **服务发现** (discover
> 内部支持模块，构建这个组件可以发现在consul中注册的Braid节点，并且它会将节点信息同步到负载均衡器中。

* **RPC** (dispatcher | register
> dispatcher
```go
disp := dispatcher.New(ConsulAddress)

// ctx 用于传递追踪数据的上下文
// targetNod 目标节点
// serviceName 服务名
// meta 用户自定义数据，不需要传nil
// body 消息体
disp.Call(ctx, targetNod, serviceName, meta, body)
```
> register
```go
// 通过 监听端口 构建register (rpc server)
reg := register.New("NodeName", WithListen(":1201"))

// 将功能函数注册到本节点
reg.Regist("serviceName", func(ctx, in[]byte) (out []byte, err error) {})

reg.Run()
defer reg.Close()

```
* **分布式追踪** (tracer
> 提供基于jaeger的分布式追踪服务，同时支持慢查询
> 即便采样率非常低，只要有调用超出设置时间 #SlowSpanLimit# #SlowRequestLimit#，这次调用也必然会被打印。
```go
// 基于 1/1000 的采样率构建 Tracer
t := tracer.New("NodeName", JaegerAddress, WithProbabilistic(0.001))
t.Init()
```

* **容器发现** (基于registerator
> 这里没有实现服务发现，而是采用了容器发现作为发现系统,
> 在Dockerfile中设置env `SERVICE_NAME` 作为节点名, `SERVICE_TAG` 作为发现标签。
```Dockerfile
ENV SERVICE_TAGS=braid,calculate
ENV SERVICE_14222_NAME=calculate
EXPOSE 14222
```

* **日志** (log


***
#### 一些完整的样例
[Gateway](https://github.com/pojol/braid-gateway "网关节点")

#### WIKI
[WIKI](https://github.com/pojol/braid/wiki "WIKI")

#### QQ Group
1057895060