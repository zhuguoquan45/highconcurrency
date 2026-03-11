TCP 的 **状态机（TCP State Machine）** 描述的是 **一个 TCP 连接从创建到关闭的整个生命周期状态变化**。
如果你在面试里讲清楚这个，基本说明你 **TCP 基础很扎实**。

先给一个面试一句话总结：

> **TCP 状态机定义了 TCP 连接在建立、通信、关闭过程中各个状态及其转换关系，总共有 11 个状态。**

---

# 1 TCP 的 11 个状态

TCP 一共有 **11 个状态**：

```id="7ofslq"
CLOSED
LISTEN
SYN_SENT
SYN_RECV
ESTABLISHED
FIN_WAIT_1
FIN_WAIT_2
CLOSE_WAIT
LAST_ACK
TIME_WAIT
CLOSING
```

简单分三类：

| 类型   | 状态                                           |
| ---- | -------------------------------------------- |
| 连接建立 | LISTEN / SYN_SENT / SYN_RECV                 |
| 数据传输 | ESTABLISHED                                  |
| 连接关闭 | FIN_WAIT / CLOSE_WAIT / LAST_ACK / TIME_WAIT |

---

# 2 TCP 状态机整体图

![Image](https://flylib.com/books/3/223/1/html/2/files/18fig12.gif)

![Image](https://www.ibm.com/support/pages/system/files/inline-images/Flow%20chart%20TCP%20connection_0.jpg)

![Image](https://www.researchgate.net/publication/221403452/figure/fig2/AS%3A668968192843783%401536505854245/TCP-state-transition-diagram.png)

![Image](https://upload.wikimedia.org/wikipedia/en/5/57/Tcp_state_diagram.png)

TCP 的状态变化主要围绕两件事：

```id="gsnnrh"
建立连接（三次握手）
关闭连接（四次挥手）
```

---

# 3 连接建立阶段

服务器启动：

```id="aovhhg"
CLOSED → LISTEN
```

等待客户端连接。

---

### 第一步

客户端发送 SYN：

```id="2chkl9"
CLOSED → SYN_SENT
```

---

### 第二步

服务器收到 SYN：

```id="l9b80n"
LISTEN → SYN_RECV
```

然后发送：

```id="l9zzoh"
SYN + ACK
```

---

### 第三步

客户端收到：

```id="zo1qfo"
SYN + ACK
```

客户端发送 ACK：

```id="4n9dh4"
SYN_SENT → ESTABLISHED
```

服务器收到 ACK：

```id="7y1vgr"
SYN_RECV → ESTABLISHED
```

连接建立。

---

# 4 数据传输阶段

双方处于：

```id="bcstbo"
ESTABLISHED
```

可以进行：

```id="uvmsbh"
双向数据传输
```

---

# 5 连接关闭阶段

关闭连接时会进入多个状态。

假设 **客户端主动关闭**。

---

## 第一步

客户端发送 FIN：

```id="dqr6ey"
ESTABLISHED → FIN_WAIT_1
```

---

## 第二步

服务器收到 FIN：

```id="1wcx5e"
ESTABLISHED → CLOSE_WAIT
```

服务器回复 ACK。

客户端：

```id="5qesoc"
FIN_WAIT_1 → FIN_WAIT_2
```

---

## 第三步

服务器数据发送完：

发送 FIN：

```id="7l5ifd"
CLOSE_WAIT → LAST_ACK
```

---

## 第四步

客户端收到 FIN：

发送 ACK：

```id="s1n3bh"
FIN_WAIT_2 → TIME_WAIT
```

服务器收到 ACK：

```id="3mp84c"
LAST_ACK → CLOSED
```

---

## 第五步

客户端等待：

```id="4x6v7h"
2MSL
```

然后：

```id="dtq0ss"
TIME_WAIT → CLOSED
```

连接彻底结束。

---

# 6 CLOSING 状态

这个状态比较少见。

出现条件：

```id="5c4zsh"
双方同时发送 FIN
```

流程：

```id="fr5evc"
FIN_WAIT_1
      ↓
   CLOSING
      ↓
  TIME_WAIT
```

这种叫：

```id="6awwi0"
simultaneous close
```

---

# 7 CLOSE_WAIT 为什么危险

很多线上服务器问题都卡在：

```id="s6if63"
CLOSE_WAIT
```

含义：

```id="x36oqo"
对方已经关闭连接
但本端还没关闭
```

常见原因：

```id="7pd3pv"
程序没有调用 close()
```

导致：

```id="u8f29b"
连接泄漏
```

---

# 8 TIME_WAIT 为什么存在

TIME_WAIT 的两个作用：

### 1 保证最后 ACK 到达

如果 ACK 丢失：

```id="3yhsn6"
Server 会重发 FIN
```

客户端还能处理。

---

### 2 防止旧数据包

等待：

```id="z9os8e"
2MSL
```

保证：

```id="18ue1s"
旧连接数据全部消失
```

---

# 9 TCP 状态机简化理解

其实只要记住这个核心流程：

```id="2zsoxt"
建立连接

CLOSED
LISTEN
SYN_SENT
SYN_RECV
ESTABLISHED
```

关闭连接：

```id="ly3buj"
ESTABLISHED
FIN_WAIT_1
FIN_WAIT_2
TIME_WAIT
CLOSED
```

服务器关闭：

```id="9o9x6p"
ESTABLISHED
CLOSE_WAIT
LAST_ACK
CLOSED
```

---

# 10 面试标准回答（40秒）

如果面试官问：

**TCP 状态机是什么？**

可以这样回答：

> TCP 状态机描述了 TCP 连接从建立到关闭的整个生命周期，一共定义了 11 个状态，包括 CLOSED、LISTEN、SYN_SENT、SYN_RECV、ESTABLISHED、FIN_WAIT_1、FIN_WAIT_2、CLOSE_WAIT、LAST_ACK、TIME_WAIT 和 CLOSING。连接建立时通过三次握手从 LISTEN 或 SYN_SENT 进入 ESTABLISHED；数据传输在 ESTABLISHED 状态完成；连接关闭时通过四次挥手进入 FIN_WAIT、CLOSE_WAIT、LAST_ACK 和 TIME_WAIT 等状态，最终回到 CLOSED。TCP 状态机保证了连接建立和关闭过程的可靠性。

---

如果你是准备 **DDoS / CDN / 网络安全 / Go 后端**岗位（从你前面的问题看就是这个方向），面试官通常还会继续追问这些 **TCP 状态机延伸题**：

* 为什么 **CLOSE_WAIT 很多说明程序有问题**
* **TIME_WAIT 太多如何优化**
* **SYN_RECV 队列是什么**
* **半连接队列 vs 全连接队列**
* **SYN Flood 如何打满状态机**

如果你愿意，我可以给你整理一张 **“TCP + DDoS 攻击对应状态机图”（非常适合面试）**，很多安全厂商都会问这个。
