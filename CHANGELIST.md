

# v1.2.23
> 重新设计mailbox的接口， 主要的理由有
1. 消息的类型 `广播消息` 和 `普通消息` 不应该由接收端指定
2. 分开 `topic` 和 `channel` 的定义（不在内部自己处理，让用户感知
3. `统一` proc & cluster 的消息发布内部逻辑
4. 去掉使用回调传递消息的方式，现在一致采用 `chan` 传输消息。

```go
    // old
    consumer := braid.Mailbox().Sub(mailbox.Proc, topic).Shared()
    consumer.OnArrived(func (msg *mailbox.Message) error {
        // todo
        return nil
    })

    braid.Mailbox().Pub("topic", &message{})

    // new
    topic := braid.Mailbox().Topic("topic")
    channel := Channel("channel", mailbox.ScopeProc)

    go func() {
        for {
            select {
                case <-channel.Arrived():
            }
        }
    }

    topic.Pub( &message{} )

```