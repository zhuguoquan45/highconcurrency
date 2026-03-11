# GMP 调度模型

## 核心概念

Go runtime 使用 **GMP 模型**实现 M:N 调度（N 个 goroutine 映射到 M 个 OS 线程）。

| 角色 | 含义 |
|------|------|
| G (Goroutine) | 用户态轻量级线程，初始栈 2KB，可动态增长 |
| M (Machine) | OS 线程，由操作系统调度 |
| P (Processor) | 逻辑处理器，持有本地 goroutine 队列，数量由 GOMAXPROCS 决定 |

---

## 调度流程

```
全局队列 (GRQ)
      ↓
  P 本地队列 (LRQ) ←→ work stealing
      ↓
  M 执行 G
```

1. G 创建后优先放入当前 P 的本地队列（LRQ，最多 256 个）
2. LRQ 满了则放入全局队列（GRQ）
3. M 绑定 P，从 LRQ 取 G 执行
4. LRQ 为空时，M 从 GRQ 或其他 P 的 LRQ 偷取（work stealing）

---

## 关键机制

### 抢占式调度
- Go 1.14 之前：协作式，goroutine 主动让出（函数调用时检查）
- Go 1.14 之后：基于信号的异步抢占（SIGURG），解决死循环不让出问题

### 系统调用处理
- G 发起阻塞系统调用时，M 与 P 解绑
- P 绑定新的 M 继续执行其他 G
- 系统调用返回后，G 尝试重新绑定 P，失败则进入全局队列

### work stealing
- P 本地队列为空时，随机从其他 P 偷取一半 G
- 保证 CPU 利用率，避免某些 P 空闲

---

## GOMAXPROCS

```go
runtime.GOMAXPROCS(n) // 设置 P 的数量，默认等于 CPU 核数
```

- CPU 密集型：设为 CPU 核数
- IO 密集型：可适当增大

---

## 面试高频问题

**Q: goroutine 和线程的区别？**
- goroutine 初始栈 2KB，线程 MB 级别
- goroutine 由 Go runtime 调度，线程由 OS 调度
- goroutine 切换在用户态，成本极低（~100ns），线程切换需内核态（~1μs）
- 单进程可轻松跑百万 goroutine

**Q: P 的数量设置多少合适？**
- 默认 = CPU 核数，适合 CPU 密集型
- IO 密集型可以更大，但 Go 的异步 IO 模型通常不需要

**Q: goroutine 阻塞时发生什么？**
- channel/mutex 阻塞：G 挂起，M 继续执行其他 G（不阻塞 M）
- 系统调用阻塞：M 与 P 解绑，P 找新 M 继续工作

**Q: 为什么 Go 能支持百万 goroutine？**
- 栈初始 2KB（线程 8MB），内存占用小
- 用户态调度，切换成本低
- work stealing 保证负载均衡
