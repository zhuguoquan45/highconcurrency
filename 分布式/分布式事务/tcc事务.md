明白了，我帮你整理一份 **条理清晰、结构完整的 TCC 事务学习文档**，从概念、流程、特性、设计、实现到协调者容错机制，都有清晰分层，适合微服务初学者系统学习。

---

# TCC（Try-Confirm-Cancel）事务学习文档

---

## 一、概念简介

**TCC（Try-Confirm-Cancel）事务** 是分布式系统中实现 **最终一致性** 的事务模型，通过 **业务层三段式控制**替代传统的两阶段提交（2PC），适合微服务架构。

* **Try 阶段**：尝试执行操作，预留资源，检查条件是否满足
* **Confirm 阶段**：提交操作，完成业务逻辑
* **Cancel 阶段**：回滚 Try 阶段的操作，释放资源

> 核心思想：在业务层控制资源，保证最终一致性，减少分布式锁等待和性能瓶颈。

---

## 二、事务流程

以 **账户转账** 为例：

1. 协调者发起事务
2. 调用 **账户A Try 扣款**
3. 调用 **账户B Try 收款**
4. 所有 Try 成功 → 执行 Confirm
5. 任一 Try 失败 → 执行 Cancel

**流程示意：**

```
协调者
   ├─> 服务A: TryDeduct
   ├─> 服务B: TryAdd
   ├─> 确认全部 Try 成功?
   │       ├─> 是 → Confirm 阶段
   │       └─> 否 → Cancel 阶段
```

---

## 三、TCC事务特性

| 特性        | 描述                          |
| --------- | --------------------------- |
| **最终一致性** | 系统可通过重试 Confirm/Cancel 达到一致 |
| **幂等性**   | Confirm/Cancel 必须可重复执行      |
| **资源预留**  | Try 阶段锁定业务资源，避免冲突           |
| **可伸缩性**  | 无需全局锁，适合微服务架构               |

---

## 四、设计要点

1. **幂等性保证**

   * Confirm 和 Cancel 必须幂等，避免重复调用产生副作用
2. **事务日志持久化**

   * 记录事务 ID、状态（TRYING/CONFIRMING/CANCELLING/DONE）、参与者状态
3. **悬挂事务处理**

   * Try 成功但未收到 Confirm/Cancel，需要定时补偿
4. **空操作处理**

   * 某些场景下 Confirm/Cancel 仅更新状态，无实际操作
5. **超时补偿机制**

   * 参与者可主动回滚长时间未完成事务，提高可用性

---

## 五、协调者崩溃处理

### 1. 问题描述

* 协调者在执行 Confirm/Cancel 过程中崩溃
* 部分参与者可能已经完成 Confirm
* 系统需保证 **最终一致性**

### 2. 解决方案

* **事务日志持久化**：记录 tx_id、状态、参与者状态
* **重启恢复机制**：

  1. 扫描未完成事务日志
  2. 根据状态补发 Confirm 或 Cancel
  3. 利用参与者幂等性保证安全
* **最终一致性**：所有参与者状态一致，事务完成

---

## 六、实现示例（Go + PostgreSQL）

### 数据库表设计

```sql
CREATE TABLE tcc_log (
    tx_id TEXT PRIMARY KEY,
    status TEXT NOT NULL,          -- TRYING / CONFIRMING / CANCELLING / DONE
    participants JSONB NOT NULL,   -- [{"service":"AccountA","status":"TryDone"}, ...]
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
```

### 协调者核心逻辑

```go
// StartTCC: 发起事务 -> Try -> Confirm/Cancel
// RecoverPendingTransactions: 重启恢复未完成事务
```

### 参与者接口示例

```go
func callTry(service, txID string) error {}
func callConfirm(service, txID string) error {} // 幂等
func callCancel(service, txID string) error {}  // 幂等
```

---

## 七、TCC vs 2PC 对比

| 特性   | 2PC              | TCC                    |
| ---- | ---------------- | ---------------------- |
| 阶段   | Prepare + Commit | Try + Confirm + Cancel |
| 阻塞   | 高（数据库锁）          | 低（业务层资源控制）             |
| 幂等性  | 不强制              | 必须                     |
| 适用场景 | 强一致性             | 微服务 / 最终一致性            |

---

## 八、总结

* **TCC适用场景**：资金、库存、订单等跨服务业务
* **核心要点**：业务层控制 + 日志持久化 + 幂等操作
* **协调者容错**：日志恢复 + 补偿重试 → 保证最终一致性
* **优缺点**：

  * ✅ 高性能、无全局锁
  * ✅ 高可用，可恢复协调者宕机
  * ❌ 开发复杂度高，需要三段逻辑和幂等设计

---

我可以在这个基础上帮你画一份 **流程图 + 协调者崩溃恢复图**，做成一页学习图册，让学习 TCC 一目了然。

你希望我画吗？
