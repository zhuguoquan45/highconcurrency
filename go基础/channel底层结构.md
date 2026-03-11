这个是 **Go 面试非常高频的一题**，很多人只知道 `channel 用来通信`，但如果能讲到 **runtime 结构 + 队列 + 加锁 + 调度器**，面试官基本会点头。

先给一句面试版总结：

> **Go 的 channel 本质是 runtime 中的 `hchan` 结构体，内部包含一个环形缓冲区（buffer）、发送队列、接收队列以及一把互斥锁，用于在 goroutine 之间安全地传递数据。**

下面拆开讲。

---

# 1 channel 在 runtime 中的结构

channel 在 Go runtime 中的结构体叫：

```go
runtime.hchan
```

简化后的核心结构大概是这样：

```go
type hchan struct {
    qcount   uint           // 当前队列中的元素数量
    dataqsiz uint           // 环形队列大小
    buf      unsafe.Pointer // 指向环形缓冲区
    elemsize uint16         // 元素大小

    sendx    uint           // 发送索引
    recvx    uint           // 接收索引

    recvq    waitq          // 接收等待队列
    sendq    waitq          // 发送等待队列

    lock     mutex          // 互斥锁
}
```

可以理解为：

```
channel
 ├── 环形缓冲区 buffer
 ├── send queue（发送阻塞队列）
 ├── recv queue（接收阻塞队列）
 └── mutex锁
```

---

# 2 环形缓冲区（buffer）

如果是 **有缓冲 channel**

```go
ch := make(chan int, 5)
```

runtime 会创建一个 **环形数组**：

```
buf:

+----+----+----+----+----+
|    |    |    |    |    |
+----+----+----+----+----+
```

两个指针：

```
sendx -> 写入位置
recvx -> 读取位置
```

示例：

```
sendx = 3
recvx = 1
```

读取顺序：

```
1 -> 2 -> 3
```

然后继续循环：

```
4 -> 0 -> 1
```

这就是 **环形队列（ring buffer）**。

优势：

* 不需要移动数组
* O(1) 入队出队

---

# 3 发送队列（sendq）

当 **channel 满了** 时：

```
buffer 已满
```

发送 goroutine 会被 **阻塞**。

runtime 会把 goroutine 封装成：

```
sudog
```

然后放入：

```
sendq
```

结构类似：

```
sendq

G1 -> G2 -> G3
```

这些 goroutine 会 **挂起**。

---

# 4 接收队列（recvq）

当 **channel 为空** 时：

```
buffer 空
```

接收 goroutine 会被 **阻塞**。

然后进入：

```
recvq
```

队列结构：

```
recvq

G4 -> G5 -> G6
```

---

# 5 send 操作流程（面试重点）

代码：

```go
ch <- value
```

runtime 大概流程：

### 第一步：加锁

```go
lock(&hchan.lock)
```

因为多个 goroutine 可能同时操作。

---

### 第二步：优先检查 recvq

如果有 goroutine 在等接收：

```
recvq != empty
```

直接：

```
sender -> receiver
```

**数据直接拷贝给接收方**。

不走 buffer。

```
G1(send) ---> G2(recv)
```

这是 **最快路径**。

---

### 第三步：写入 buffer

如果：

```
buffer 未满
```

写入：

```
buf[sendx] = value
sendx++
qcount++
```

---

### 第四步：阻塞

如果：

```
buffer 满
```

当前 goroutine：

```
进入 sendq
```

然后：

```
gopark()
```

挂起 goroutine。

---

# 6 recv 操作流程

代码：

```go
v := <-ch
```

流程类似。

---

### 第一步：加锁

```
lock
```

---

### 第二步：检查 sendq

如果有发送者在等待：

```
sendq != empty
```

直接：

```
sender -> receiver
```

同时唤醒 sender。

---

### 第三步：读取 buffer

如果：

```
buffer 有数据
```

执行：

```
value = buf[recvx]
recvx++
qcount--
```

---

### 第四步：阻塞

如果：

```
buffer 空
```

goroutine：

```
进入 recvq
```

然后：

```
gopark()
```

挂起。

---

# 7 sudog 结构

等待队列里不是直接存 goroutine，而是：

```
sudog
```

结构类似：

```go
type sudog struct {
    g *g
    elem unsafe.Pointer
    next *sudog
}
```

作用：

```
goroutine + 数据地址
```

用于在唤醒时传递数据。

---

# 8 channel 的锁

channel 有一把：

```
mutex
```

原因：

因为可能出现：

```
多个 goroutine 同时

send
recv
close
```

需要保护：

```
buffer
queue
index
```

---

# 9 无缓冲 channel（重要）

如果：

```go
ch := make(chan int)
```

则：

```
dataqsiz = 0
```

没有 buffer。

通信模式变成：

```
send 必须等待 recv
recv 必须等待 send
```

流程：

```
G1 send
      ↓
   数据拷贝
      ↓
G2 recv
```

这叫：

> **同步通信（同步 channel）**

---

# 10 一张图总结 channel

```
           +-------------------+
           |      hchan        |
           |-------------------|
           | qcount            |
           | dataqsiz          |
           | sendx / recvx     |
           |-------------------|
           | buffer (ring)     |
           |-------------------|
           | sendq (等待发送)   |
           | recvq (等待接收)   |
           |-------------------|
           | mutex lock        |
           +-------------------+
```

---

# 面试标准答案（30秒版）

如果面试官问：

**channel 底层结构是什么？**

可以这样回答：

> Go 的 channel 在 runtime 中实现为 `hchan` 结构体，内部包含一个环形缓冲区 buffer、发送等待队列 sendq、接收等待队列 recvq 以及一把互斥锁。
> 当 goroutine 发送数据时，如果有接收者在 recvq 中等待，会直接把数据拷贝给接收者；如果没有，则写入 buffer；如果 buffer 已满，发送 goroutine 会被挂起并进入 sendq。
> 接收操作也是类似逻辑。
> 通过这种设计，Go 实现了 goroutine 之间高效、安全的通信。

---

如果你准备 **Go 中高级面试**，下一题面试官 **100% 会继续追问**：

1️⃣ **channel 为什么是线程安全的？**
2️⃣ **close(channel) 底层发生了什么？**
3️⃣ **select 是怎么实现的？**（超级高频）
4️⃣ **为什么 channel 不会忙等？**
5️⃣ **channel 会不会发生内存拷贝？**

我可以给你整理一份 **Go 并发 20 个最狠面试题（带 runtime 级答案）**。很多都是一线大厂会问的。
