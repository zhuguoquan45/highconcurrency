下面给你整理了一份 **基于本次会话内容的系统性总结文档**。
内容围绕我们讨论的核心：**分布式事务 2PC（Two-Phase Commit）机制、问题以及工程实践方案**。这份文档我按照 **技术文档结构**整理，方便你复习、面试或者整理知识体系。

---

# 分布式事务 2PC 机制总结文档

## 1. 分布式事务背景

在单体数据库中，一个事务可以依赖数据库的 **ACID特性**：

* **Atomicity（原子性）**
* **Consistency（一致性）**
* **Isolation（隔离性）**
* **Durability（持久性）**

但在 **微服务架构** 中，一个业务操作可能涉及多个数据库，例如：

```
订单服务
支付服务
库存服务
```

例如下单流程：

```
创建订单
扣减库存
扣减余额
```

如果其中一个操作失败，就会导致 **数据不一致问题**。

因此需要 **分布式事务机制**。

---

# 2. Two-Phase Commit（2PC）

2PC 是经典的 **分布式事务协议**，通过 **两个阶段** 来保证事务一致性。

系统角色：

```
Transaction Coordinator（事务协调者）
Participants（事务参与者）
```

结构：

```
         Coordinator
             │
      ┌──────┼──────┐
      │      │      │
     DB1    DB2    DB3
```

---

# 3. 第一阶段：Prepare Phase

协调者向所有节点发送：

```
PREPARE
```

参与者执行本地事务：

```sql
BEGIN;
UPDATE account SET balance = balance - 100 WHERE id = 1;
PREPARE TRANSACTION 'tx1';
```

执行结果：

| 节点  | 状态       |
| --- | -------- |
| DB1 | PREPARED |
| DB2 | PREPARED |
| DB3 | PREPARED |

此时特点：

* 事务修改 **已经写入 WAL**
* 数据 **未提交**
* **锁仍然持有**
* 等待协调者决议

可以查询：

```sql
SELECT * FROM pg_prepared_xacts;
```

---

# 4. 第二阶段：Commit Phase

如果所有节点都返回成功：

协调者发送：

```
COMMIT PREPARED 'tx1'
```

参与者提交事务。

最终：

```
事务正式提交
锁释放
```

如果有节点失败：

```
ROLLBACK PREPARED 'tx1'
```

事务回滚。

---

# 5. Commit 阶段部分成功问题

可能发生：

```
DB1 commit 成功
DB2 commit 成功
DB3 commit 失败
```

状态：

| 节点  | 状态  |
| --- | --- |
| DB1 | 已提交 |
| DB2 | 已提交 |
| DB3 | 未提交 |

导致：

```
数据不一致
```

例如：

```
账户A -100
账户B 未 +100
```

---

# 6. 2PC 的恢复机制

为了避免上述问题，协调者必须 **持久化决议日志**：

```
tx1 -> COMMIT
```

恢复流程：

```
Coordinator crash
↓
读取事务日志
↓
继续发送 COMMIT
```

因此：

```
commit 决议不可逆
```

系统会 **不断重试 commit**。

---

# 7. 节点宕机恢复机制

如果某个节点宕机：

```
DB3 crash
```

重启后数据库执行：

```
WAL recovery
```

恢复：

```
PREPARED transaction
```

例如：

```
tx1
```

数据库状态：

```
prepared but not committed
```

此时数据库：

```
等待 commit 或 rollback 指令
```

协调者会通过 **重试机制** 再次发送：

```
COMMIT PREPARED 'tx1'
```

最终提交。

---

# 8. Coordinator 宕机问题（2PC阻塞）

如果协调者崩溃：

```
Coordinator crash
```

所有节点状态：

```
PREPARED
```

但没有人发送：

```
COMMIT
或
ROLLBACK
```

结果：

```
事务阻塞
锁一直持有
```

数据库可能卡住。

这就是：

**2PC Blocking Problem**

---

# 9. “节点询问 Coordinator”机制

很多资料会说：

```
节点恢复后会询问 coordinator
```

实际上：

数据库本身 **不会主动询问**。

真实机制是：

### 协调者重试

```
Coordinator 持久化 commit log
↓
不断 retry commit
↓
节点恢复后执行 commit prepared
```

在 **XA协议** 中才有：

```
XA RECOVER
```

获取未决事务：

```
in-doubt transaction
```

然后事务管理器决定：

```
xa_commit
xa_rollback
```

---

# 10. 2PC 的核心问题

2PC 在互联网系统中很少直接使用，原因：

### 1 阻塞问题

PREPARED 状态：

```
锁不会释放
```

可能导致系统卡死。

---

### 2 协调者单点

```
Coordinator crash
```

事务全部阻塞。

---

### 3 网络分区

节点无法确定：

```
commit
or
rollback
```

---

### 4 性能问题

2PC 需要：

```
2 次网络通信
磁盘日志
锁持有
```

吞吐量低。

---

# 11. 互联网常用替代方案

## 1 Saga（最常见）

核心思想：

```
本地事务 + 补偿事务
```

流程：

```
订单创建
支付
库存
```

失败：

```
补偿操作
```

例如：

```
退款
```

优点：

```
高可用
无锁
```

缺点：

```
业务复杂
```

---

## 2 TCC

三阶段：

```
Try
Confirm
Cancel
```

例子：

```
Try: 冻结余额
Confirm: 扣款
Cancel: 解冻
```

特点：

```
强一致
需要业务实现
```

---

## 3 本地事务 + MQ

最常见模式：

```
DB transaction
+
Message Queue
```

流程：

```
数据库事务提交
发送消息
消费者执行
```

保证：

```
最终一致性
```

---

## 4 事务消息

例如：

```
RocketMQ Transaction Message
```

流程：

```
Half Message
Commit Message
```

---

# 12. 大规模分布式数据库方案

现代分布式数据库通常采用：

```
2PC + 共识协议
```

例如：

### Google Spanner

使用：

```
2PC
+
Paxos
+
TrueTime
```

特点：

```
Coordinator 可复制
避免单点
保证一致性
```

---

# 13. 总结

2PC 的核心思想：

```
Prepare
Commit
```

优点：

```
强一致性
实现简单
```

缺点：

```
阻塞
协调者单点
性能差
```

因此在现代微服务架构中：

```
很少直接使用 2PC
```

更常见的是：

```
Saga
TCC
消息事务
```

实现：

```
最终一致性
```

---

如果你愿意，我可以再帮你整理一份 **更高级的版本**（很多架构师用来复习的那种）：

**《分布式事务完整知识体系》**

会包含：

* 2PC / 3PC 深度原理
* CAP 与分布式事务关系
* Saga / TCC / Seata 实现
* Spanner / TiDB / CockroachDB 事务机制
* 分布式事务面试 50 题

那份基本可以当 **架构级分布式事务笔记**。
