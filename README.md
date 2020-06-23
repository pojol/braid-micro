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
> braid对外提供的接口

``` go
/module
    /* 选举，提供接口 IsMaster 给用户判断当前节点是否为主节点。 在任意时候都会只存在一个主节点，当原有的主节点下线后，会选举出新的主节点*/
    /elector

    /* 基于grpc的远程调用，提供client 和 server 端的支持 */
    /rpc

    /* 分布式追踪，支持各种行为追踪，grpc，redis，http，慢查询，等 */
    /tracer
```
![image.png](https://i.loli.net/2020/06/19/CwbvuhyjKkXLf6d.png)


#### 插件（plug-in
> braid中各种组件实现采用的策略，也支持用户引入自定义的插件。

```go
/plugin

    /* 负载均衡器插件接口文件 */
    /balancer
        /* SmoothWeightRoundrobin 平滑加权轮询实现 */
        /swrrbalancer

    /* 发现器接口文件 */
    /discover
        /* 基于consul的服务发现实现 */
        /consuldiscover

    /* 选举器接口文件 */
    /elector
        /* 基于consul的选举实现 */
        /consulelector

    /* 链接器接口文件 */
    / linker
        /* 基于redis的连接器实现 */
        /redislinker

```

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