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

#### 组件
* **选举** (election
```go
// 通过 "邮件服务节点名" "Consul" "LockTick 选举竞争频率" 构建election
elec := election.New("mail", ConsulAddress, WithLockTick(1000))

elec.Run()
defer elec.Close()

// 获取当前节点是否为Master节点（Master节点只会存在一个，且当Master节点下线后会从其他同名节点中选举出新的Master.
elec.IsMaster()
```

* **RPC** 
> client 通过传入`目标节点`（节点名）信息，调用负载均衡器选择一个权重较轻的节点进行发送（默认采用`平滑加权轮询`
> client 会自动`发现`注册到braid的节点。
```go
rc := client.New("base", consulAddr, client.WithTracing())
rc.Discover()
defer rc.Close()

conn, err := client.GetConn("mail") // 获取一个邮件节点的连接
if err != nil {
    goto EXT
}
defer conn.Put()    // 还给池

cc = pbraid.NewCalculateClient(conn.ClientConn)
res, err = cc.Addition(ctx.Request().Context(), &pbraid.AddReq{})
if err != nil {
    conn.Unhealthy()    // 如果调用失败，将连接设置为不健康的，由池进行销毁。
}
```
> server
```go
type mailServer struct {
	pbraid.MailServer
}

// SendMail 发送邮件服务
func (cs *mailServer) SendMail(ctx context.Context, req *SendMailReq) (*SendMailRes, error) {
	return res, nil
}

s := server.New("mail", server.WithListen(":14222"), server.WithTracing())
pbraid.RegisterMailServer(server.Get(), &mailServer{})

s.Run()
defer s.Close()

```
* **分布式追踪** (tracer
> 提供基于jaeger的分布式追踪服务，同时支持慢查询
> 即便采样率非常低，只要有调用超出设置时间 #SlowSpanLimit# #SlowRequestLimit#，这次调用也必然会被打印。
```go
// 基于 1/1000 的采样率构建 Tracer
t := tracer.New("mail", JaegerAddress, WithProbabilistic(0.001))
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
> 日志模块基于zap，提供默认日志构建以及多种可选的日志
```go
log.New(log.Config{mode, path, suffex})

// 这里多添加了 系统诊断日志模块 以及 用户行为日志
log.New(log.Config{mode, path, suffex}, 
    WithSys(log.Config{}), 
    WithBehavior(log.Config{}))

// 普通日志使用 zap.sugared 灵活使用
log.Debugf("%v\n", "")
// 结构化日志需要参照 log_sys.go 进行自定义输出
log.SysError("module", "func", desc)
```

***

#### WIKI
[WIKI](https://github.com/pojol/braid/wiki "WIKI")