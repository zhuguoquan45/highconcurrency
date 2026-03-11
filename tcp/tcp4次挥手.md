TCP 断开连接用的是 **四次挥手（Four-Way Handshake）**。核心原因一句话：

> **TCP 是全双工协议，双方都需要分别关闭发送方向，所以需要 4 次报文。**

下面一步一步看。

---

## 1 四次挥手整体流程

![Image](https://media.licdn.com/dms/image/v2/D4D12AQHlJizxB7DTaA/article-inline_image-shrink_1000_1488/article-inline_image-shrink_1000_1488/0/1670060130972?e=2147483647\&t=W5Pcw8-gnGNC4691g2io1ycHbrr9zxL1ILFqN6SF_R8\&v=beta)

![Image](https://cdn.shopify.com/s/files/1/0615/5911/1885/files/1_94222b65-b1be-4221-8c1f-a37a93549023.png?v=1748512416)

![Image](https://i.sstatic.net/JjMcf.png)

![Image](https://media.licdn.com/dms/image/v2/D4D12AQFkL24S0-qIww/article-inline_image-shrink_400_744/article-inline_image-shrink_400_744/0/1670060039953?e=2147483647\&t=jL9NuzXc6-DeV9Qn-junpyAisV5J_ComOSYNvy9xH6Y\&v=beta)

假设：

* **Client（客户端）**主动关闭
* **Server（服务器）**被动关闭

流程如下：

```
Client                        Server

FIN ------------------------>
      (1)

     <---------------------- ACK
             (2)

     <---------------------- FIN
             (3)

ACK ------------------------>
      (4)
```

---

# 2 第一次挥手（FIN）

客户端发送：

```
FIN = 1
seq = x
```

表示：

```
我没有数据要发了
```

客户端状态：

```
FIN_WAIT_1
```

---

# 3 第二次挥手（ACK）

服务器收到 FIN 后回复：

```
ACK = 1
ack = x + 1
```

表示：

```
我知道你要关闭发送了
```

服务器状态：

```
CLOSE_WAIT
```

客户端状态变成：

```
FIN_WAIT_2
```

⚠️ 这里注意：

服务器 **可能还有数据没发完**，所以不会马上发 FIN。

---

# 4 第三次挥手（FIN）

服务器数据发送完后：

```
FIN = 1
seq = y
```

表示：

```
我也没有数据发了
```

服务器状态：

```
LAST_ACK
```

---

# 5 第四次挥手（ACK）

客户端收到 FIN 后：

```
ACK = 1
ack = y + 1
```

客户端进入：

```
TIME_WAIT
```

服务器收到 ACK：

```
CLOSED
```

连接关闭。

---

# 6 为什么需要四次？

很多面试官喜欢问这个。

原因：

```
TCP 是全双工通信
```

双方发送通道 **独立关闭**。

流程：

```
A 关闭 A→B
B 确认
B 关闭 B→A
A 确认
```

所以需要：

```
FIN
ACK
FIN
ACK
```

---

# 7 为什么 TIME_WAIT 要等 2MSL？

客户端最后会进入：

```
TIME_WAIT
```

时间：

```
2MSL
```

MSL：

```
Maximum Segment Lifetime
报文最大生存时间
```

等待原因：

### 1 防止最后 ACK 丢失

如果 ACK 丢失：

```
Server 会重发 FIN
```

客户端还在 TIME_WAIT：

```
可以重新 ACK
```

---

### 2 防止旧连接数据干扰新连接

等待：

```
旧数据全部消失
```

避免：

```
旧包进入新连接
```

---

# 8 TCP 状态变化总结

客户端：

```
ESTABLISHED
FIN_WAIT_1
FIN_WAIT_2
TIME_WAIT
CLOSED
```

服务器：

```
ESTABLISHED
CLOSE_WAIT
LAST_ACK
CLOSED
```

---

# 9 一张状态流图

```
Client                          Server

ESTABLISHED                     ESTABLISHED
     │                               │
     │ FIN                           │
     ▼                               │
FIN_WAIT_1                           │
     │                               │
     │ ACK                           ▼
     ▼                           CLOSE_WAIT
FIN_WAIT_2                           │
                                     │ FIN
                                     ▼
                                 LAST_ACK
     ▲                               │
     │ ACK                           │
     │                               │
TIME_WAIT                            ▼
     │                             CLOSED
     ▼
CLOSED
```

---

# 10 面试最标准回答（30秒）

如果面试官问：

**TCP 四次挥手过程？**

可以这样答：

> TCP 连接断开需要四次挥手。首先主动关闭方发送 FIN 表示没有数据发送了，进入 FIN_WAIT_1 状态；对方收到后回复 ACK，进入 CLOSE_WAIT 状态，主动关闭方进入 FIN_WAIT_2；当被动关闭方数据发送完后发送 FIN，进入 LAST_ACK；最后主动关闭方回复 ACK 并进入 TIME_WAIT 状态，等待 2MSL 后彻底关闭连接。由于 TCP 是全双工协议，双方需要分别关闭发送方向，因此需要四次挥手。

---

💡 面试官 **90% 会继续追问这几个**：

1️⃣ **为什么 TIME_WAIT 是 2MSL？**
2️⃣ **CLOSE_WAIT 太多说明什么问题？**
3️⃣ **TIME_WAIT 太多怎么优化？**
4️⃣ **为什么建立连接是三次握手而不是四次？**
5️⃣ **SYN Flood 和三次握手关系？**

如果你在准备 **网络 / CDN / DDoS / Go 后端岗位**，我可以给你整理一套 **TCP 高频面试 40 题（带标准答案）**，很多是云厂商和安全厂商非常爱问的。
