我给你整理一份 **完整、系统、适合学习和面试的 MQ事务（Message Queue Transaction）综合文档**。
内容包含：

1. MQ事务产生背景
2. 为什么会出现不一致
3. MQ事务解决方案
4. RocketMQ事务消息原理
5. Go实现示例
6. 本地消息表（主流方案）
7. 工程实践建议
8. 面试总结

---

# 一、MQ事务是什么

**MQ事务（Message Queue Transaction）** 是一种解决 **数据库操作和消息发送一致性问题** 的机制。

在微服务架构中，服务之间通常通过 MQ 进行通信，例如：

```
订单服务 → MQ → 库存服务
```

典型流程：

```
用户下单
   ↓
订单服务写数据库
   ↓
发送MQ消息
   ↓
库存服务消费消息扣库存
```

问题就在于：

```
数据库操作 和 MQ发送
属于两个不同系统
```

它们 **不在同一个事务中**。

因此可能出现数据不一致。

---

# 二、为什么会出现数据不一致

如果没有事务控制，可能出现两种问题：

## 1 数据库成功，消息发送失败

流程：

```
1 数据库 commit 成功
2 MQ发送失败
```

结果：

```
订单存在
库存没有扣
```

原因可能是：

```
MQ网络异常
MQ broker宕机
程序崩溃
发送超时
```

---

## 2 消息发送成功，数据库失败

流程：

```
1 MQ发送成功
2 数据库事务回滚
```

结果：

```
库存扣减
订单不存在
```

原因：

```
SQL错误
唯一索引冲突
数据库宕机
程序崩溃
```

---

# 三、MQ事务解决目标

MQ事务要保证：

```
本地事务成功 → 消息一定发送
本地事务失败 → 消息一定不发送
```

最终实现：

```
微服务数据最终一致性
```

---

# 四、MQ事务解决方案

工程中常见三种方案：

| 方案     | 一致性  | 使用情况 |
| ------ | ---- | ---- |
| MQ事务消息 | 最终一致 | 常用   |
| 本地消息表  | 最终一致 | ⭐最主流 |
| 2PC    | 强一致  | 很少   |

互联网系统 **80%~90% 使用本地消息表**。

---

# 五、RocketMQ事务消息原理

RocketMQ提供 **Transactional Message**。

核心思想：

```
半消息 + 本地事务 + 回查机制
```

---

## 事务消息流程

流程如下：

```
Producer
   |
   | 1 发送半消息
   v
Broker
   |
   | 2 执行本地事务
   v
Producer
   |
   | 3 commit / rollback
   v
Broker
   |
   | 4 投递消息
   v
Consumer
```

---

## 半消息（Prepare Message）

Producer先发送一条：

```
prepare message
```

特点：

```
Broker会存储
但不会投递给消费者
```

---

## 执行本地事务

Producer执行：

```
数据库事务
```

例如：

```
create order
```

---

## 提交事务

事务完成后：

```
commit → MQ投递消息
rollback → 删除消息
```

---

## 事务回查机制

如果出现：

```
Producer崩溃
commit未发送
```

Broker会回查：

```
CheckLocalTransaction
```

询问：

```
事务状态？
commit
rollback
unknown
```

---

# 六、Go实现事务消息示例

使用：

```
rocketmq-client-go
```

安装：

```
go get github.com/apache/rocketmq-client-go/v2
```

---

## 事务监听器

```go
type OrderTransactionListener struct{}

func (l *OrderTransactionListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {

    orderId := string(msg.Body)

    err := CreateOrder(orderId)

    if err != nil {
        return primitive.RollbackMessageState
    }

    return primitive.CommitMessageState
}

func (l *OrderTransactionListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {

    orderId := string(msg.Body)

    if OrderExists(orderId) {
        return primitive.CommitMessageState
    }

    return primitive.RollbackMessageState
}
```

---

## Producer发送事务消息

```go
producer, _ := rocketmq.NewTransactionProducer(
    listener,
    rocketmq.WithNameServer([]string{"127.0.0.1:9876"}),
)

producer.Start()

msg := primitive.NewMessage(
    "order_topic",
    []byte("order123"),
)

producer.SendMessageInTransaction(
    context.Background(),
    msg,
    nil,
)
```

---

# 七、本地消息表（Outbox Pattern）

在真实公司中，**更常见的是本地消息表方案**。

核心思想：

```
业务数据 + 消息
写入同一个数据库事务
```

---

## 表结构

订单表：

```
orders
------

id
user_id
status
```

消息表：

```
message_outbox
--------------

id
topic
payload
status
retry_count
```

---

## 写入流程

```
BEGIN

insert order
insert message_outbox

COMMIT
```

保证：

```
订单存在
消息一定存在
```

---

## 后台任务发送MQ

定时任务：

```
扫描 message_outbox
```

发送消息：

```
mq.send()
```

成功：

```
status = sent
```

失败：

```
retry++
```

---

# 八、消费端必须幂等

MQ系统可能：

```
重复投递
```

因此消费端必须：

```
幂等处理
```

常见做法：

### 方法1 去重表

```
message_consume_log
```

记录：

```
message_id
```

---

### 方法2 业务唯一ID

例如：

```
order_id
```

如果已经处理：

```
直接忽略
```

---

# 九、MQ事务架构图

完整架构通常是：

```
订单服务
   |
   | 本地事务
   v
数据库
   |
   | message_outbox
   v
定时任务
   |
   v
MQ
   |
   v
库存服务
   |
   v
扣库存
```

---

# 十、工程实践建议

实际系统推荐：

### 1 优先使用本地消息表

优点：

```
稳定
不依赖MQ特性
易扩展
```

---

### 2 消费端必须幂等

避免：

```
重复消费
```

---

### 3 消息必须可重试

避免：

```
消息丢失
```

---

### 4 增加死信队列

处理：

```
多次失败消息
```

---

# 十一、MQ事务 vs 其他分布式事务

| 方案   | 一致性  | 复杂度 | 场景   |
| ---- | ---- | --- | ---- |
| 2PC  | 强一致  | 高   | 金融   |
| 3PC  | 强一致  | 高   | 理论   |
| TCC  | 强一致  | 高   | 核心交易 |
| Saga | 最终一致 | 中   | 微服务  |
| MQ事务 | 最终一致 | 低   | ⭐最常见 |

---

# 十二、面试总结（标准答案）

MQ事务用于解决：

```
数据库操作 与 MQ消息发送 的一致性问题
```

可能出现：

```
数据库成功消息失败
消息成功数据库失败
```

解决方案：

```
MQ事务消息
本地消息表（主流）
```

RocketMQ事务消息机制：

```
1 发送半消息
2 执行本地事务
3 commit / rollback
4 Broker事务回查
```

保证：

```
最终一致性
```

---

如果你愿意，我可以再给你一份 **《分布式事务完整技术体系文档》**（非常适合面试和系统设计），里面会把：

```
2PC
3PC
TCC
Saga
MQ事务
本地消息表
```

**用一张架构图全部讲清楚。**
很多大厂面试其实就是围绕这张图展开的。
