这个问题其实是 **Go 高并发服务器面试的经典题**。先给一个面试一句话版本：

> **Go 能实现百万连接，核心依赖于 goroutine 的轻量级调度（GMP 模型）、网络 I/O 的 epoll/kqueue 事件驱动、goroutine 的小栈和动态扩展，以及 runtime 的 netpoll 网络轮询机制。**

简单说就是：

```
goroutine + epoll + netpoll + GMP 调度
```

下面把关键机制拆开讲清楚。

---

# 1 传统线程模型为什么做不到百万连接

如果用传统线程：

```
1 连接 = 1 线程
```

假设：

```
100万连接
```

Linux 线程栈默认：

```
8MB
```

内存需求：

```
8MB × 1000000 = 8TB
```

直接爆炸。

而且线程上下文切换成本很高。

---

# 2 Go 的 goroutine 模型

Go 的模型：

```
1 连接 = 1 goroutine
```

goroutine 初始栈：

```
2KB
```

假设：

```
100万 goroutine
```

内存：

```
2KB × 1000000 ≈ 2GB
```

是可以接受的。

而且栈会 **动态增长**：

```
2KB → 4KB → 8KB → ...
```

不用时还会收缩。

---

# 3 Go 网络 I/O 的关键：netpoll

Go 网络库不是简单的 blocking IO。

底层使用 **网络轮询器**：

```
netpoll
```

不同系统对应：

| OS      | 机制     |
| ------- | ------ |
| Linux   | epoll  |
| Mac     | kqueue |
| Windows | IOCP   |

也就是说：

```
goroutine
   ↓
netpoll
   ↓
epoll
```

---

# 4 Go 网络阻塞的真相

看代码：

```go
conn.Read(buf)
```

看起来像阻塞。

但实际上：

```
goroutine 阻塞
线程不会阻塞
```

流程是：

```
goroutine 调用 read
        ↓
fd 注册到 epoll
        ↓
goroutine 挂起
        ↓
线程继续执行其他 goroutine
```

当 socket 可读：

```
epoll_wait
      ↓
netpoll 返回事件
      ↓
唤醒 goroutine
```

---

# 5 Go 网络调度流程

完整流程：

```
goroutine Read
      │
      ▼
runtime.netpoll
      │
      ▼
epoll_wait
      │
      ▼
socket ready
      │
      ▼
netpoll 返回
      │
      ▼
goroutine runnable
      │
      ▼
GMP 调度执行
```

---

# 6 GMP 调度模型

Go 的调度模型：

```
G = goroutine
M = OS thread
P = processor
```

关系：

```
G → P → M → CPU
```

特点：

* goroutine 数量远大于线程
* runtime 负责调度
* 支持 work stealing

例如：

```
1M goroutine
1000 threads
```

也是正常的。

---

# 7 Go 网络连接结构

Go 的连接对象：

```
net.TCPConn
```

内部包含：

```
pollDesc
```

结构：

```
TCPConn
   │
   ▼
netFD
   │
   ▼
pollDesc
   │
   ▼
epoll fd
```

pollDesc 用于：

```
注册 epoll
唤醒 goroutine
```

---

# 8 典型百万连接服务器写法

经典 Go TCP 服务器：

```go
ln, _ := net.Listen("tcp", ":8080")

for {
    conn, _ := ln.Accept()

    go handle(conn)
}
```

处理函数：

```go
func handle(conn net.Conn) {
    buf := make([]byte, 4096)

    for {
        n, err := conn.Read(buf)
        if err != nil {
            return
        }

        conn.Write(buf[:n])
    }
}
```

特点：

```
1 连接 = 1 goroutine
```

因为 goroutine 很轻，所以可以支持：

```
100万连接
```

---

# 9 Go 支持百万连接的核心原因

核心四点：

### 1 goroutine 轻量级

```
2KB stack
```

支持百万级。

---

### 2 GMP 调度

用户态调度：

```
低上下文切换
```

---

### 3 netpoll

使用：

```
epoll / kqueue
```

避免：

```
线程阻塞
```

---

### 4 非阻塞网络

goroutine 阻塞：

```
线程不会阻塞
```

---

# 10 实际百万连接的系统架构

真实生产环境通常这样：

```
           ┌───────────────┐
           │   LoadBalancer │
           └───────┬───────┘
                   │
           ┌───────▼────────┐
           │  Go Server     │
           │                │
           │ 1M goroutine   │
           │                │
           │ netpoll        │
           │ epoll          │
           └────────────────┘
```

同时还会调优：

```
ulimit -n
somaxconn
tcp_tw_reuse
tcp_fin_timeout
```

---

# 11 百万连接的 Linux 调优

常见调优：

### 文件描述符

```
ulimit -n 2000000
```

---

### backlog

```
net.core.somaxconn = 65535
```

---

### 端口范围

```
net.ipv4.ip_local_port_range
```

---

### TIME_WAIT

```
tcp_tw_reuse
```

---

# 12 面试 30 秒回答版本

如果面试官问：

**Go 如何实现百万连接？**

可以这样回答：

> Go 支持百万连接主要依赖三个机制。首先是 goroutine 的轻量级设计，每个 goroutine 初始栈只有 2KB，可以支持百万级并发。其次是 Go runtime 的 GMP 调度模型，可以在少量线程上调度大量 goroutine。最后是 Go 的网络 I/O 使用 netpoll 机制，底层基于 epoll/kqueue 实现事件驱动，当 goroutine 进行网络读写时，如果 socket 未就绪，goroutine 会被挂起并注册到 epoll，线程可以继续执行其他 goroutine。当 epoll 返回事件时再唤醒对应 goroutine。因此 Go 可以用少量线程处理大量网络连接。

---

如果你在准备 **Go 后端 / CDN / DDoS / 网关类岗位**（你前面问了很多网络相关问题，其实非常契合），面试官通常还会继续追问：

1️⃣ Go netpoll 的实现原理
2️⃣ epoll LT / ET 在 Go 中如何使用
3️⃣ Go 如何避免 thundering herd
4️⃣ Go HTTP 服务器为什么性能高
5️⃣ Go netpoll 和 Nginx epoll 有什么区别

如果你愿意，我可以给你整理一份 **《Go 高并发服务器面试 20 题（非常硬核）》**，很多都是一线大厂会问的。
