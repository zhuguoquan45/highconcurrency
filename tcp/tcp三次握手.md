TCP 建立连接需要 **三次握手（Three-Way Handshake）**。先给一个面试一句话版：

> **TCP 三次握手的目的是让客户端和服务器确认双方的发送能力和接收能力正常，并同步初始序列号（ISN）。**

---

## 1 三次握手流程

![Image](https://static.afteracademy.com/images/what-is-a-tcp-3-way-handshake-process-three-way-handshaking-establishing-connection-6a724e77ba96e241.jpg)

![Image](https://www.tutorialspoint.com/assets/questions/media/59897/10%28a%29.jpg)

![Image](https://scaler.com/topics/images/steps-of-a-3-way-handshake-for-establishing-the-connection.webp)

![Image](https://www.tcpipguide.com/free/diagrams/tcpopen3way.png)

流程如下：

```id="31p7sq"
Client                        Server

SYN ------------------------->
      (1)

     <---------------------- SYN + ACK
             (2)

ACK ------------------------>
      (3)
```

---

# 2 第一次握手（SYN）

客户端发送：

```id="r10wxy"
SYN = 1
seq = x
```

含义：

```id="sw10f6"
客户端请求建立连接
```

客户端状态变为：

```id="k9feha"
SYN_SENT
```

服务器收到后进入：

```id="70mta6"
SYN_RECV
```

---

# 3 第二次握手（SYN + ACK）

服务器回应：

```id="0m6mty"
SYN = 1
ACK = 1
seq = y
ack = x + 1
```

含义：

```id="52z35c"
确认收到客户端 SYN
并发送自己的 SYN
```

---

# 4 第三次握手（ACK）

客户端收到后发送：

```id="oq9qtv"
ACK = 1
ack = y + 1
```

此时：

```id="kbspn9"
客户端 -> ESTABLISHED
服务器 -> ESTABLISHED
```

连接建立完成。

---

# 5 TCP 状态变化

客户端：

```id="tk7hql"
CLOSED
SYN_SENT
ESTABLISHED
```

服务器：

```id="5xulz6"
LISTEN
SYN_RECV
ESTABLISHED
```

---

# 6 为什么是三次而不是两次？

这是面试官非常爱问的。

核心原因：

```id="u8rjo9"
防止旧的 SYN 报文造成错误连接
```

举个例子：

如果只有 **两次握手**：

1. 客户端发送 SYN
2. 服务器回复 SYN+ACK
3. 连接建立

问题：

如果 **旧的 SYN 报文延迟到达服务器**：

```id="tw0vui"
Server 以为客户端要建立连接
```

服务器就会：

```id="3gtuh7"
建立无效连接
```

但客户端根本没有发起连接。

三次握手解决方法：

```id="u2v0h8"
第三次 ACK
```

如果客户端不确认：

```id="os9ztk"
服务器不会建立连接
```

---

# 7 三次握手的真正目的

其实不只是建立连接。

还有两个重要目的：

### 1 确认双方收发能力

```id="50g0ql"
Client -> Server 能发送
Server -> Client 能发送
双方都能接收
```

---

### 2 同步初始序列号

TCP 需要：

```id="q57tkn"
seq number
```

防止：

```id="djnthg"
旧数据包
```

干扰新连接。

---

# 8 SYN Flood 和三次握手

很多安全岗位会问。

攻击流程：

```id="nkn0uk"
攻击者发送 SYN
但不发送 ACK
```

服务器会：

```id="xbte35"
SYN_RECV
```

维护：

```id="v4gqtr"
半连接队列
```

如果大量 SYN：

```id="o4sbsr"
半连接队列被打满
```

服务器无法接受新连接。

常见防护：

```id="7i70aa"
SYN Cookie
SYN Proxy
SYN Rate Limit
```

---

# 9 三次握手流程图

```id="vdv5sf"
Client                               Server

CLOSED                               LISTEN
   │                                     │
   │ SYN                                 │
   ▼                                     │
SYN_SENT                                 │
                                         ▼
                                     SYN_RECV
   │                                     │
   │ ACK                                 │
   ▼                                     ▼
ESTABLISHED                         ESTABLISHED
```

---

# 10 面试标准回答（30秒）

如果面试官问：

**TCP 三次握手过程？**

可以这样回答：

> TCP 建立连接需要三次握手。首先客户端发送 SYN 报文请求建立连接，并携带初始序列号 seq=x，进入 SYN_SENT 状态；服务器收到后回复 SYN+ACK 报文，确认号为 ack=x+1，并发送自己的序列号 seq=y，进入 SYN_RECV 状态；客户端收到后发送 ACK 报文 ack=y+1，双方进入 ESTABLISHED 状态，连接建立完成。三次握手的主要目的是确认双方的收发能力，并同步初始序列号，同时避免历史连接请求造成错误连接。

---

💡 如果你在准备 **网络 / CDN / DDoS / Go 后端**面试（你前面问的基本都是这类岗位），面试官接下来 **几乎一定会追问**：

1️⃣ **为什么不是四次握手？**
2️⃣ **SYN Flood 原理？**
3️⃣ **SYN Cookie 怎么实现？**
4️⃣ **半连接队列和全连接队列？**
5️⃣ **TCP Fast Open 是什么？**

如果你愿意，我可以给你整理一份 **《TCP + DDoS 面试 60 题知识图谱》**，基本覆盖安全厂商、云厂商面试常问内容。
