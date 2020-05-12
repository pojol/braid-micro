## braid
轻量的微服务框架，通过compose.yml文件可以将braid提供的微服务组件轻易聚合到server上。

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/41686ce5edf844fc8b81cffd13cc0550)](https://app.codacy.com/manual/pojol/braid?utm_source=github.com&utm_medium=referral&utm_content=pojol/braid&utm_campaign=Badge_Grade_Dashboard)
[![drone](http://47.96.147.176:8001/api/badges/pojol/braid/status.svg?branch=develop)](dev)
[![codecov](https://codecov.io/gh/pojol/braid/branch/master/graph/badge.svg)](https://codecov.io/gh/pojol/braid)

> `注:`当前v1.1.x版本为`原型`版本 


#### 组件
* **Election** 选举组件, 提供自动选举功能
    > 在运行期自动进行选举策略，保证在多个同名节点中有一个主节点和多个从节点
    > 当节点因为任何原因不在提供服务后，会自动选举出新的主节点。

* **Register** 功能注册
    > 主要提供Regist方法，将用户定义的实现聚合到服务中心。

* **Rpc** 远程调用
    > 当服务函数通过接口braid.Regist进行过注册，则可以通过call进行远程调用
    > 访问基于rpc，并且在braid中它将被自动的负载均衡到同名节点中
    > 外部或者内部的程序只需按指定的url规则既可进行访问 braid.Call("/login/guest", in) out

* **Tracer** 分布式追踪
    > 提供基于jaeger的分布式追踪服务，同时支持慢查询
    > 即便采样率非常低，只要有调用超出设置时间 #SlowSpanLimit# #SlowRequestLimit#，这次调用也必然会被打印。

* **Logger** 日志模块
    > 日志


#### Quack Start
> 编写一个简易的login服务
> `braid_compose.yml`
```yaml
name : gateway
mode : debug
tracing : true

# 安装模块列表
install :
  - logger
  - election
  - rpc
  - tracer

# 配置模块参数，不设置使用default参数.
config :
  logger_path : /var/log/gateway
  logger_suffex : .sys
```

***
#### 一些完整的样例
[Gateway](https://github.com/pojol/braid-gateway "网关节点")


