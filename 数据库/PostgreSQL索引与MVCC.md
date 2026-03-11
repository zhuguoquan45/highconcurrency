# PostgreSQL 索引与 MVCC

## 索引原理

### B-Tree 结构（默认索引）

PostgreSQL 默认使用 **B-Tree** 索引：
- 所有数据存在叶子节点，叶子节点通过双向链表相连
- 非叶子节点只存 key，层数少，IO 次数少
- 支持等值查询、范围查询、排序

```
         [30]
        /    \
    [10,20]  [40,50]
   /  |  \   /  |  \
 [数据行链表...]
```

### 索引类型对比

| 类型 | 适用场景 |
|------|---------|
| B-Tree | 等值、范围、排序（默认，覆盖 90% 场景） |
| Hash | 仅等值查询，不支持范围 |
| GIN | 数组、JSONB、全文检索（多值列） |
| GiST | 地理空间、范围类型 |
| BRIN | 超大表按物理顺序存储的列（如时间戳），极小体积 |

```sql
-- 普通 B-Tree
CREATE INDEX idx_user_id ON orders(user_id);

-- JSONB 用 GIN
CREATE INDEX idx_data ON events USING GIN(data);

-- 全文检索
CREATE INDEX idx_content ON articles USING GIN(to_tsvector('english', content));

-- 时间序列大表用 BRIN
CREATE INDEX idx_created ON logs USING BRIN(created_at);
```

### 覆盖索引（Index-Only Scan）

```sql
CREATE INDEX idx_name_age ON users(name, age);

-- Index-Only Scan，不回表
SELECT name, age FROM users WHERE name = 'Alice';

-- 需要回表（Heap Fetch）
SELECT * FROM users WHERE name = 'Alice';
```

### 联合索引最左前缀

```sql
CREATE INDEX idx_a_b_c ON t(a, b, c);

-- 能用索引
WHERE a = 1
WHERE a = 1 AND b = 2
WHERE a = 1 AND b = 2 AND c = 3
WHERE a = 1 AND b > 2  -- a 用索引，b 范围后 c 失效

-- 不能用索引
WHERE b = 2            -- 跳过 a
WHERE b = 2 AND c = 3  -- 跳过 a
```

### 部分索引（Partial Index）

```sql
-- 只索引未完成的订单，体积小、效率高
CREATE INDEX idx_pending ON orders(user_id) WHERE status = 'pending';
```

---

## 事务隔离级别

PostgreSQL 支持 4 种隔离级别，**默认 READ COMMITTED**：

| 级别 | 脏读 | 不可重复读 | 幻读 | 实现 |
|------|------|-----------|------|------|
| READ UNCOMMITTED | ✗（降级为 RC） | ✓ | ✓ | MVCC |
| READ COMMITTED | ✗ | ✓ | ✓ | MVCC |
| REPEATABLE READ | ✗ | ✗ | ✗（PG 用 MVCC 解决）| Snapshot Isolation |
| SERIALIZABLE | ✗ | ✗ | ✗ | SSI |

**与 MySQL 的关键区别：**
- PG 的 REPEATABLE READ 通过 MVCC 快照解决幻读，**不需要 Gap Lock**
- PG 没有 Gap Lock
- MySQL InnoDB 默认 REPEATABLE READ，PG 默认 READ COMMITTED

```sql
BEGIN ISOLATION LEVEL REPEATABLE READ;
-- 或
SET TRANSACTION ISOLATION LEVEL SERIALIZABLE;
SHOW transaction_isolation;
```

---

## MVCC（多版本并发控制）

### 核心思想
读不加锁，通过行版本实现快照读，解决读写冲突。

### 实现机制

**每行隐藏字段：**
- `xmin`：插入该行的事务 ID（行何时变得可见）
- `xmax`：删除/更新该行的事务 ID（行何时失效）

**版本链（通过 xmin/xmax）：**
```
UPDATE 操作 = 旧行标记 xmax + 新行写入 xmin
旧行不立即删除，由 VACUUM 清理（dead tuple）
```

**Snapshot 可见性判断：**
```
xmin 已提交 且 xmin <= snapshot_xmin  → 可见
xmax 未提交 或 xmax > snapshot_xmin   → 当前版本有效
```

### 快照读 vs 当前读

```sql
-- 快照读（MVCC，不加锁）
SELECT * FROM t WHERE id = 1;

-- 当前读（读最新版本，加锁）
SELECT * FROM t WHERE id = 1 FOR UPDATE;
SELECT * FROM t WHERE id = 1 FOR SHARE;
INSERT / UPDATE / DELETE
```

### VACUUM 与 dead tuple

MVCC 的代价：UPDATE/DELETE 产生旧版本行（dead tuple），需要 VACUUM 清理：

```sql
-- 手动清理
VACUUM orders;

-- 重建表（锁表，慎用）
VACUUM FULL orders;

-- 查看 dead tuple 数量
SELECT relname, n_dead_tup, n_live_tup
FROM pg_stat_user_tables
WHERE relname = 'orders';
```

---

## 锁机制

### 行锁

```sql
-- 排他锁（写锁）
SELECT * FROM t WHERE id = 1 FOR UPDATE;

-- 共享锁（读锁）
SELECT * FROM t WHERE id = 1 FOR SHARE;

-- 跳过已锁定行（高并发队列消费）
SELECT * FROM tasks WHERE status = 'pending'
ORDER BY id LIMIT 10
FOR UPDATE SKIP LOCKED;

-- 锁不到立即报错
SELECT * FROM t WHERE id = 1 FOR UPDATE NOWAIT;
```

### Advisory Lock（应用级锁）

```sql
-- 获取会话级锁（事务结束不自动释放）
SELECT pg_advisory_lock(12345);
SELECT pg_advisory_unlock(12345);

-- 获取事务级锁（事务结束自动释放）
SELECT pg_advisory_xact_lock(12345);
```

适用场景：分布式任务调度、防止重复执行的定时任务。

### 查看锁冲突

```sql
-- 查看当前锁等待
SELECT pid, query, wait_event_type, wait_event
FROM pg_stat_activity
WHERE wait_event_type = 'Lock';

-- 查看锁详情
SELECT * FROM pg_locks WHERE NOT granted;
```

---

## EXPLAIN 分析

```sql
EXPLAIN ANALYZE SELECT * FROM orders WHERE user_id = 100 AND status = 'paid';
```

**关键节点类型（从差到好）：**

| 节点 | 说明 |
|------|------|
| Seq Scan | 全表扫描，大表需优化 |
| Index Scan | 索引扫描 + 回表 |
| Index Only Scan | 覆盖索引，不回表，最优 |
| Bitmap Heap Scan | 多条件索引合并，中等 |

**关注指标：**
- `cost`：估算代价（启动代价..总代价）
- `rows`：预估行数
- `actual time`：实际执行时间
- `Buffers`：缓存命中情况

```sql
-- 查看是否走了索引
EXPLAIN (ANALYZE, BUFFERS) SELECT ...;
```

---

## 面试高频问题

**Q: PG 为什么不需要 Gap Lock 也能防幻读？**
- PG 的 REPEATABLE READ 在事务开始时创建快照，整个事务只看快照内的数据
- 其他事务新插入的行 xmin > 当前快照，对当前事务不可见
- 不需要锁住间隙，MVCC 天然解决

**Q: PG 和 MySQL 的 MVCC 实现有何不同？**
| 对比 | PostgreSQL | MySQL InnoDB |
|------|-----------|-------------|
| 版本存储 | 行内 xmin/xmax，旧版本在原表 | undo log 链表 |
| 清理机制 | VACUUM 清理 dead tuple | purge 线程清理 undo log |
| 默认隔离级别 | READ COMMITTED | REPEATABLE READ |
| RR 防幻读 | MVCC 快照 | Gap Lock |

**Q: 什么情况下索引失效？**
- 对索引列做函数：`WHERE date_trunc('year', created_at) = '2024-01-01'`（改用范围查询）
- 隐式类型转换：`WHERE phone = 13800138000`（phone 是 text）
- 联合索引不满足最左前缀
- 统计信息过旧导致规划器选错（执行 `ANALYZE table` 更新统计信息）

**Q: VACUUM 和 VACUUM FULL 的区别？**
- VACUUM：并发执行，不锁表，标记 dead tuple 空间可复用，但不归还给 OS
- VACUUM FULL：锁表，重建表文件，归还磁盘空间，代价大，生产慎用
- autovacuum 自动触发普通 VACUUM

**Q: SERIALIZABLE 隔离级别有什么代价？**
- 使用 SSI（Serializable Snapshot Isolation），检测读写依赖冲突
- 可能在 COMMIT 时报错：`could not serialize access due to read/write dependencies`
- 应用层必须实现重试逻辑
- 并发性能低于 REPEATABLE READ
