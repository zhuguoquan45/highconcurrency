这个问题在 **TCP / 高并发服务器 / DDoS 面试**里非常常见。很多人只知道 `listen(backlog)`，但其实 **Linux 有两个队列**，这是面试加分点。

先给一句面试版总结：

> **backlog 用于限制服务器在三次握手过程中维护的连接队列长度，主要涉及半连接队列（SYN Queue）和全连接队列（Accept Queue），用于缓存尚未被应用程序 accept 的连接。**

---

# 1 backlog 在哪里使用

服务器启动 TCP 监听时：

```c
listen(fd, backlog);
```

例如：

```go
ln, _ := net.Listen("tcp", ":8080")
```

底层最终也会调用：

```text
listen(fd, backlog)
```

`backlog` 的作用就是：

```text
限制等待被 accept 的连接数量
```

---

# 2 Linux 实际有两个队列（重点）

TCP 连接建立过程中，Linux 内核维护 **两个队列**：

```text
1 SYN队列（半连接队列）
2 Accept队列（全连接队列）
```

整体流程：

```text
Client ---- SYN ----> Server
                ↓
           SYN Queue
                ↓
Client <--- SYN+ACK ---
                ↓
Client ---- ACK ----> Server
                ↓
           Accept Queue
                ↓
             accept()
```

---

# 3 SYN 队列（半连接队列）

当服务器收到：

```text
SYN
```

连接还没有完成三次握手。

此时连接状态：

```text
SYN_RECV
```

连接会进入：

```text
SYN Queue
```

也叫：

```text
半连接队列
```

如果这个队列满了：

```text
新的 SYN 会被丢弃
```

这就是：

```text
SYN Flood 攻击
```

---

# 4 Accept 队列（全连接队列）

当三次握手完成：

```text
ACK 到达
```

连接进入：

```text
ESTABLISHED
```

然后进入：

```text
Accept Queue
```

等待应用程序调用：

```c
accept()
```

取走连接。

如果 **Accept Queue 满了**：

```text
新的连接会被拒绝
```

客户端看到：

```text
connection refused
```

---

# 5 backlog 控制哪个队列？

这是面试最容易问的点。

在 **Linux 2.2 之后**：

```text
backlog 控制 Accept Queue
```

而：

```text
SYN Queue 大小
```

由系统参数控制：

```text
net.ipv4.tcp_max_syn_backlog
```

---

# 6 队列关系图

```text
                SYN
Client --------------------> Server

                        ┌──────────────┐
                        │  SYN Queue   │
                        │ 半连接队列    │
                        └──────┬───────┘
                               │
                         SYN+ACK
                               │
                               ▼
                         Client ACK
                               │
                               ▼
                        ┌──────────────┐
                        │ Accept Queue │
                        │ 全连接队列    │
                        └──────┬───────┘
                               │
                           accept()
                               │
                               ▼
                          应用程序
```

---

# 7 队列满了会发生什么

### SYN 队列满

表现：

```text
SYN 丢弃
```

客户端：

```text
连接重试
```

服务器：

```text
SYN_RECV 很多
```

典型：

```text
SYN Flood
```

---

### Accept 队列满

表现：

```text
连接建立但无法 accept
```

客户端：

```text
connection reset
或超时
```

服务器：

```text
ESTABLISHED 堆积
```

原因通常是：

```text
程序 accept 太慢
```

---

# 8 backlog 实际大小

实际 backlog 不是完全等于参数值。

内核会取：

```text
min(backlog, somaxconn)
```

Linux 默认：

```text
net.core.somaxconn = 4096
```

例如：

```c
listen(fd, 100000)
```

最终：

```text
4096
```

---

# 9 高并发服务器常见调优

例如百万连接服务器：

```text
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
```

同时：

```text
ulimit -n
```

提高文件描述符。

---

# 10 面试标准回答（40秒）

如果面试官问：

**backlog 队列的作用是什么？**

可以这样回答：

> backlog 用于限制 TCP 服务器在连接建立过程中维护的连接队列长度。Linux 内核中实际上存在两个队列：半连接队列（SYN Queue）和全连接队列（Accept Queue）。当服务器收到 SYN 时，连接会进入 SYN 队列，状态为 SYN_RECV；当三次握手完成后，连接进入 Accept 队列，等待应用程序调用 accept 取走。backlog 参数主要限制 Accept Queue 的长度，如果队列满了，新连接可能会被拒绝或丢弃。

---

💡 在 **网络 / CDN / DDoS / Go 服务端**面试里，这个问题后面 **90% 会继续追问**：

1️⃣ **半连接队列 vs 全连接队列区别**
2️⃣ **SYN Flood 如何打满 SYN Queue**
3️⃣ **SYN Cookie 如何绕过 SYN Queue**
4️⃣ **为什么服务器会出现大量 SYN_RECV**
5️⃣ **accept 队列满了会发生什么**

如果你愿意，我可以给你画一张 **“TCP 三次握手 + backlog + SYN Flood 攻击”完整流程图**，这个在安全面试里非常好用。
