## braid
轻量的微服务框架，通过compose.yml文件可以将braid提供的微服务组件轻易聚合到server上。

[![drone](http://47.96.147.176:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)

> `注:`当前v1.1.x版本为`原型`版本 


#### 组件
* **Service** 服务组件，提供服务的注册和路由功能
    
    > 用户只需调用braid.Regist接口即可将自己提供的服务注册到中心
    
* **Election** 选举组件, 提供自动选举功能
    > 在运行期自动进行选举策略，保证在多个同名节点中有一个主节点和多个从节点
    > 当节点因为任何原因不在提供服务后，会自动选举出新的主节点。

* **Linker** 链接组件，维护各个节点之间的链接关系
    > 当这个组件生效后，在进行服务查找时，调用关系将被缓存。在第二次执行同样的请求时不需要再通过服务查找。
    > 且在服务下线或者不可用时，链接将被自动移除。
    
* **Caller** 远程调用
    > 当服务函数通过接口braid.Regist进行过注册，则可以通过call进行远程调用
    > 访问基于rpc，并且在braid中它将被自动的负载均衡到同名节点中
    > 外部或者内部的程序只需按指定的url规则既可进行访问 braid.Call("/login/guest", in) out

* **Tracer** 分布式追踪
    > 提供基于jaeger的分布式追踪服务，同时支持慢查询
    > 即便采样率非常低，只要有调用超出设置时间 #SlowSpanLimit# #SlowRequestLimit#，这次调用也必然会被打印。


#### Quack Start
> 编写一个简易的login服务
> `braid_compose.yml`
```yaml
name : login
mode : debug
tracing : false

depend :
    consul_addr : http://127.0.0.1:8900

install : 
    log :
        open : true
        path : /var/log/login
        suffex : .sys
    service :
        open : true
```
> `main.go`
```go
composeFile, err := ioutil.ReadFile("braid_compose.yml")
if err != nil {
    fmt.Println(err)
}

conf := &braid.ComposeConf{}
err = yaml.Unmarshal([]byte(composeFile), conf)
if err != nil {
    log.Fatalln(err)
}

err = braid.Compose(*conf, braid.DependConf{
		Consul: consulAddr,
		Redis:  redisAddr,
		Jaeger: jaegerAddr,
	})
if err != nil {
    log.Fatalln(err)
}

braid.Regist("guest", func(ctx context.Context, in []byte) (out []byte, err error) {
    // login.guest
    return nil, nil
})
braid.Run()

ch := make(chan os.Signal)
signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
<-ch

braid.Close()
```

***
#### 一些完整的样例
[Coordinate](https://github.com/pojol/braid-coordinate "协调节点")
[Gateway](https://github.com/pojol/braid-gateway "网关节点")


