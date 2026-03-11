`select` 是 Go 并发里一个很优雅但实现挺“硬核”的东西。先给一个**面试一句话版**：

> **Go 的 select 在 runtime 中通过 `selectgo()` 实现，会把每个 case 转换成一个 `scase` 结构，先随机打乱顺序做一次非阻塞检查，如果没有可执行的 case，就把当前 goroutine 注册到所有 channel 的等待队列中并挂起，等其中一个 channel 就绪时再被唤醒执行对应 case。**

下面我们拆开讲清楚它的底层机制。

---

# 1 select 在 runtime 中的入口

Go 编译器遇到：

```go
select {
case v := <-ch1:
    ...
case ch2 <- x:
    ...
default:
    ...
}
```

会编译成 runtime 调用：

```go
runtime.selectgo()
```

核心代码在：

```
src/runtime/select.go
```

主要函数：

```
selectgo()
```

---

# 2 select 的内部结构

每一个 case 在 runtime 中都会变成一个结构：

```go
type scase struct {
    c    *hchan         // channel
    elem unsafe.Pointer // 数据指针
    kind uint16         // send or recv
}
```

例如：

```go
select {
case x := <-ch1:
case ch2 <- v:
}
```

会变成类似：

```
scase[0] -> recv ch1
scase[1] -> send ch2
```

---

# 3 select 执行流程（核心）

select 的执行流程其实是 **三阶段算法**：

```
1 随机轮询检查
2 尝试执行
3 阻塞等待
```

我们详细看。

---

# 4 第一阶段：随机化 case 顺序

runtime 会 **随机打乱 case 顺序**：

```go
randperm()
```

例如：

原来：

```
case1
case2
case3
```

随机后：

```
case3
case1
case2
```

为什么？

防止：

```
case1 永远优先
```

否则：

```
select {
case <-ch1:
case <-ch2:
}
```

`ch1` 会一直被优先执行。

所以 Go 做了 **公平性设计**。

---

# 5 第二阶段：非阻塞检查

runtime 会依次检查每个 case 是否 **立即可执行**。

例如：

### recv case

检查：

```
channel 是否有数据
```

等价于：

```
qcount > 0
```

或者：

```
sendq 有 goroutine
```

如果成立：

```
直接执行
```

---

### send case

检查：

```
buffer 是否未满
```

或者：

```
recvq 有 goroutine
```

如果成立：

```
直接执行
```

---

如果某个 case 可以执行：

```
select 立即返回
```

不会阻塞。

---

# 6 第三阶段：default 判断

如果所有 case 都不能执行：

* 有 `default`

```
直接执行 default
```

* 没有 `default`

进入 **阻塞阶段**

---

# 7 第四阶段：注册等待队列

如果要阻塞：

runtime 会把当前 goroutine：

```
注册到所有 channel 的等待队列
```

例如：

```
select {
case <-ch1
case <-ch2
case <-ch3
}
```

当前 goroutine 会被注册到：

```
ch1.recvq
ch2.recvq
ch3.recvq
```

结构类似：

```
G1 -> sudog -> ch1
G1 -> sudog -> ch2
G1 -> sudog -> ch3
```

每个 channel 都有一个 **sudog 结构**。

---

# 8 goroutine 挂起

注册完后：

```
gopark()
```

当前 goroutine 被挂起。

调度器去执行其他 goroutine。

---

# 9 唤醒机制

假设：

```
ch2 收到数据
```

runtime 在 channel send 时：

会检查：

```
recvq
```

发现：

```
G1 在 select 等待
```

于是：

```
唤醒 G1
```

但有一个关键问题：

> G1 同时注册在多个 channel 上

所以 runtime 要做：

```
清理其他 channel 的等待节点
```

只保留触发的 case。

---

# 10 数据拷贝

如果是：

```
recv case
```

runtime 会：

```
sender -> receiver
```

拷贝数据。

如果是：

```
send case
```

则：

```
receiver <- value
```

---

# 11 select 的重要特性

### 1 随机公平

因为：

```
case 顺序随机
```

避免饥饿。

---

### 2 不会 busy loop

select 阻塞时：

```
gopark()
```

不会：

```
CPU 空转
```

---

### 3 O(n) 扫描

select 每次执行：

```
遍历所有 case
```

复杂度：

```
O(n)
```

所以 Go 官方建议：

```
select case 不要太多
```

---

# 12 一个典型流程图

```
select start
      │
      ▼
随机打乱 case
      │
      ▼
遍历 case 检查是否可执行
      │
      ├─ 有可执行
      │      │
      │      ▼
      │   执行 case
      │
      ▼
无可执行
      │
      ├─ 有 default
      │      │
      │      ▼
      │   执行 default
      │
      ▼
无 default
      │
      ▼
注册到 channel 队列
      │
      ▼
gopark 挂起
      │
      ▼
channel 就绪
      │
      ▼
唤醒 goroutine
      │
      ▼
执行对应 case
```

---

# 13 面试 30 秒回答模板

如果面试官问：

**select 的实现原理是什么？**

可以这样说：

> Go 的 select 在 runtime 中通过 `selectgo()` 实现。编译器会把每个 case 转换为 `scase` 结构。执行时 runtime 会先随机打乱 case 顺序，然后进行一次非阻塞检查，如果某个 case 可以立即执行就直接执行。如果没有可执行 case 且存在 default，则执行 default；如果没有 default，则把当前 goroutine 注册到所有 channel 的等待队列中并挂起。当其中某个 channel 就绪时，runtime 会唤醒 goroutine 并执行对应 case，同时清理其他 channel 上的等待节点。

---

# 14 面试官接下来 90% 会追问的

如果你答完 select，通常会继续问：

### 1

**select 为什么要随机？**

（避免 case 饥饿）

---

### 2

**select 是线程安全的吗？**

（channel 内部有 mutex）

---

### 3

**select 可以监听多少 channel？**

（理论无限，但 O(n)）

---

### 4

**select 能实现优先级吗？**

（Go 官方：不能保证）

---

### 5（高阶）

**select 为什么不会死锁？**

---

如果你愿意，我可以给你整理一套 **Go runtime 并发面试最狠 12 题**（基本是阿里 / 字节 / 腾讯 Go 岗常问的），比如：

* GMP 调度器为什么设计成这样
* work stealing 原理
* goroutine 抢占式调度实现
* channel close 底层发生什么
* map 为什么并发不安全

这套基本就是 **Go 面试核心杀手题**。
