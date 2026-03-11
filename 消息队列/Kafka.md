# Kafka 核心原理

## 基本架构

```
Producer → Broker Cluster → Consumer Group
                ↓
           ZooKeeper / KRaft (元数据管理)
```

| 概念 | 说明 |
|------|------|
| Topic | 消息分类，逻辑概念 |
| Partition | Topic 的物理分片，有序、不可变日志 |
| Broker | Kafka 服务节点 |
| Producer | 消息生产者 |
| Consumer Group | 消费者组，组内每个 Partition 只被一个 Consumer 消费 |
| Offset | 消息在 Partition 中的位置，Consumer 自己维护 |

---

## 为什么 Kafka 高吞吐

1. **顺序写磁盘**：Partition 是追加写，顺序 IO 比随机 IO 快 100 倍
2. **零拷贝（sendfile）**：数据从磁盘直接到网卡，不经过用户态
3. **批量发送**：Producer 批量压缩发送，减少网络开销
4. **分区并行**：多 Partition 并行消费，水平扩展

---

## 消息可靠性

### Producer 端
```go
// acks 配置
acks=0   // 不等确认，最快但可能丢失
acks=1   // Leader 写入即确认，Leader 宕机可能丢失
acks=-1  // 所有 ISR 副本写入才确认，最安全
```

### Broker 端
- **副本机制**：每个 Partition 有多个副本（Leader + Follower）
- **ISR（In-Sync Replicas）**：与 Leader 保持同步的副本集合
- **min.insync.replicas**：最少同步副本数，配合 acks=-1 保证强一致

### Consumer 端
- **手动提交 offset**：处理完消息后再提交，避免消息丢失
- **at-least-once**：可能重复消费，业务需幂等处理

---

## 消息顺序性

- **Partition 内有序**，Partition 间无序
- 需要全局有序：只用 1 个 Partition（牺牲并发）
- 需要局部有序：同一 key 的消息路由到同一 Partition

```go
// 指定 key，保证同 key 消息有序
producer.Send(&sarama.ProducerMessage{
    Topic: "orders",
    Key:   sarama.StringEncoder(orderID),
    Value: sarama.StringEncoder(data),
})
```

---

## 消费者组 Rebalance

**触发条件：**
- Consumer 加入或离开 Consumer Group
- Topic 的 Partition 数量变化
- Consumer 心跳超时

**影响：**
- Rebalance 期间所有 Consumer 停止消费（STW）
- 频繁 Rebalance 会影响吞吐量

**优化：**
- 增大 `session.timeout.ms` 和 `heartbeat.interval.ms`
- 使用 `CooperativeSticky` 分配策略（增量 Rebalance）

---

## 常见问题

### 消息积压
- 原因：消费速度 < 生产速度
- 解决：增加 Consumer 数量（不超过 Partition 数）、优化消费逻辑、增加 Partition

### 重复消费
- 原因：Consumer 处理完但 offset 未提交就崩溃
- 解决：业务幂等（数据库唯一键、Redis SETNX）

### 消息丢失
- Producer：acks=-1 + 重试
- Broker：副本数 >= 2，min.insync.replicas >= 2
- Consumer：手动提交 offset

---

## 面试高频问题

**Q: Kafka 和 RocketMQ 的区别？**
| 对比 | Kafka | RocketMQ |
|------|-------|----------|
| 吞吐量 | 极高（百万/s） | 高（十万/s） |
| 延迟 | ms 级 | ms 级 |
| 顺序消息 | Partition 内有序 | 支持全局/局部有序 |
| 事务消息 | 不支持 | 支持 |
| 延迟消息 | 不支持 | 支持 |
| 适用场景 | 日志、大数据 | 电商、金融 |

**Q: Kafka 如何保证消息不丢失？**
- Producer：acks=-1，retries > 0
- Broker：replication.factor >= 2，min.insync.replicas >= 2
- Consumer：手动提交 offset，处理完再提交

**Q: Kafka 的 offset 存在哪里？**
- 旧版：ZooKeeper
- 新版（0.9+）：`__consumer_offsets` 内部 Topic

**Q: 一个 Consumer Group 最多几个 Consumer 有效？**
- 等于 Partition 数量，超出的 Consumer 空闲
