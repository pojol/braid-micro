# v1.2.26
1. 将主要的注释翻译成英文
2. 为 discoverconsul 模块添加自动的不健康节点排除功能
3. 为 grpc client & server 添加 AppendInterceptors Option方法

# v1.2.25
1. 统一所有module的builder接口
2. 将 Mailbox 更名为 Pubsub （更贴近逻辑的名字
3. 调整一下构建阶段的代码（原来有些凌乱

# v1.2.24
> 1.2.23 版本补充
1. 从 codecov 切换到 coveralls.io
2. 将默认的 channel 切换为 Unbounded channel 用于避免消费过快导致阻塞。
3. 添加一些接口中的注释

# v1.2.23
> 重新设计mailbox的接口， 主要的理由有
1. 消息的类型 `广播消息` 和 `普通消息` 不应该由接收端指定
2. 分开 `topic` 和 `channel` 的定义（不在内部自己处理，让用户感知
3. `统一` proc & cluster 的消息发布内部逻辑（都采用nsq的消息发布-订阅模型

### 几个主要的概念
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


### 样例

```go
    // old
    consumer := braid.Mailbox().Sub(mailbox.Proc, topic).Shared()
    consumer.OnArrived(func (msg *pubsub.Message) error {
        // todo
        return nil
    })

    braid.Mailbox().Pub("topic", &message{})

    // new
    braid.Mailbox().RegistTopic("topic name", mailbox.ScopeProc)
    
    topic := braid.Mailbox().GetTopic("topic name")
    consumer := topic.Sub("channel name")
    consumer.Arrived(func(msg *pubsub.Message){
        // todo
    })

    topic.Pub( &message{} )

```