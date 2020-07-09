## braid
轻量的微服务框架，提供常用的组件，使用braid将使我们专注在实现上，而不需要关心主从，添加删除服务，调度，负载均衡等微服务逻辑。
另外用户也可以实现自己的插件，自定义模块的内在逻辑。

[![Go Report Card](https://goreportcard.com/badge/github.com/pojol/braid)](https://goreportcard.com/report/github.com/pojol/braid)
[![drone](http://123.207.198.57:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)

<img src="https://i.postimg.cc/B6b6CMjM/image.png" width="600">


#### Feature

###### **RPC**
> 基于grpc的远程调用，内含`服务发现` `负载均衡` `链路缓存` 等插件，基于`标签`和`名字`进行服务发现（自动获取远程目标地址，通过装配插件，可以实现各种不同的远程调用内在逻辑。（同时也支持用户自定义

###### **Discover** (服务发现
> 服务发现插件主要用于发现在系统内的节点变化，提供节点的进入，离开，以及权重值的变化等信息。

###### **Linker** (链路缓存
> 链路缓存插件，实现链路的缓存记录，通过链路缓存可以固定用户身份对节点调用的链路，辅助用户在逻辑节点中实现一些可靠的缓存操作。另外也可以提供节点当前的链路情况，可以用于权重的参考。

###### **Balancer** (负载均衡
> 负载均衡主要提供client在远程调用时的目标选取

###### **Elector** (选举
> 这里实现的选举器主要提供 IsMaster 接口，当 Master 节点离线时，系统会自动选举出新的 Master ，并且能够保证在系统中当前只会有唯一一个 Master 节点。

###### **Tracing** (分布式追踪
> 基于jaeger以及opentracing实现的分布式追踪服务，除了原有的功能外，还添加了`慢查询`的支持，在追踪系统中通过`限定值`输出慢日志，即便在采样率1/1000的设定下依然会打印命中限定的日志。



#### rpc-client sample
```go
	b := New("test")
	b.RegistPlugin(DiscoverByConsul(mock.ConsulAddr, consuldiscover.WithInterval(time.Second*3)),
		BalancerBySwrr(),
		RPCClient(grpcclient.WithPoolCapacity(128)))

	b.Run()
	defer b.Close()

	Client().Invoke(context.TODO(), "targeNodeName", "/proto.node/method", "", nil, nil)
```



#### 快速开始

> 获取braid

```bash
go get github.com/pojol/braid@latest
```

> hello,braid

* 基于Docker进行构架

  https://github.com/pojol/braid/wiki/quick-start-with-docker

* 基于Kubernetes进行构架 (v1.4.x

  https://github.com/pojol/braid/wiki/quick-start-with-k8s



#### Wiki
https://github.com/pojol/braid/wiki