## braid
轻量的微服务框架，提供常用的组件，使用braid将使我们专注在实现上，而不需要关心主从，添加删除服务，调度，负载均衡等微服务逻辑。

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid)](https://goreportcard.com/report/github.com/pojol/braid)
[![drone](http://123.207.198.57:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)


<img src="https://i.postimg.cc/B6b6CMjM/image.png" width="600">

> `注:`当前v1.1.x版本为`原型`版本 

> 获取braid

```bash
go get github.com/pojol/braid@latest
```

#### 组件 (module
> braid对外提供的组件目录

* **rpc** 远程调用
    * `client` 提供 **GetConn** 方法，通过`节点名`自动挑选一个连接。
    
    ```go
    Invoke(context.TODO(), "mail", "/bproto.listen/routing", &bproto.RouteReq{
		Nod:     nodeName,
		Service: serviceName,
		ReqBody: []byte{},
	}, res)
    ```

    * `server` grpc server的包装
    
    ```go
    s := server.New("mail", server.WithListen(":14222"), server.WithTracing())
    pbraid.RegisterMailServer(server.Get(), &mailServer{})

    s.Run()
    defer s.Close()
    ```
* **tracer** 分布式链路追踪组件
> 提供基于jaeger的分布式追踪服务，同时支持慢查询
> 即便采样率非常低，只要有调用超出设置时间 #SlowSpanLimit# #SlowRequestLimit#，这次调用也必然会被打印。

```go
// 基于 1/1000 的采样率构建 Tracer
t := tracer.New("mail", JaegerAddress, WithProbabilistic(0.001))
t.Init()
```

![image.png](https://i.loli.net/2020/06/19/CwbvuhyjKkXLf6d.png)

* **election** 选举组件
> 获取当前节点是否为Master节点（Master节点只会存在一个，且当Master节点下线后会从其他同名节点中选举出新的Master

```go
// 构建，选择一个选举器
e := election.GetBuilder(ElectionName).Build(Cfg{
    Address:           mock.ConsulAddr,
    Name:              "test",
    LockTick:          time.Second,
    RefushSessionTick: time.Second,
})


e.Run()
e.IsMaster()    // 获取当前节点是否为主节点。
e.Close()
```

#### 插件 (plugin
> 提供组件的不同实现的目录，另外也支持用户在外部自己实现plugin注册到braid.

* consul_discover 基于consul实现的服务发现&注册
* consul_election 基于consul实现的election
* balancer_swrr 平滑加权轮询


#### 其他

* **容器注册** (基于registerator
> 这里没有实现服务注册，而是采用了容器注册作为注册系统,
> 在Dockerfile中设置env `SERVICE_NAME` 作为节点名, `SERVICE_TAG` 作为注册标签。
```Dockerfile
ENV SERVICE_TAGS=braid,calculate
ENV SERVICE_14222_NAME=calculate
EXPOSE 14222
```
> 启动容器后，容器中的服务会自动注册到braid.

***

#### WIKI
[WIKI](https://github.com/pojol/braid/wiki "WIKI")