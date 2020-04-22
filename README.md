## braid
轻量的微服务框架，通过compose.yml文件可以将braid提供的微服务组件轻易聚合到server上。

> `注:`当前v1.1.x版本为`原型`版本 

#### 组件
* **Service** 服务组件，提供服务的注册和路由功能
    > 用户只需调用braid.Regist接口即可将自己提供的服务注册到中心
* **Election** 选举组件, 提供自动选举功能
    > 在运行期自动进行选举策略，保证在多个同名节点中有一个主节点和多个从节点
* **Linker** 链接组件，维护各个节点之间的链接关系
    > 当这个组件生效后，在进行服务查找时，调用关系将被缓存。在第二次执行同样的请求时不需要再通过服务查找。
* **Caller** 远程调用,rpc
    > 当服务通过接口braid.Regist进行过注册，则可以通过call进行远程调用
    > 访问基于rpc，并且在braid中它将被自动的负载均衡到同名节点中
    > 外部或者内部的程序只需按指定的url规则既可进行访问 braid.Call("/login/guest")
* **Tracer** 分布式追踪
    > 提供基于jaeger的分布式追踪服务，同时支持慢查询
    > 即便采样率非常低，braid也会将超出设置调用时间的过程打印 #SlowSpanLimit# #SlowRequestLimit#

#### Quack Start
> 一个简易的login服务
> `braid.yml`
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

	err = braid.Compose(*conf)
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