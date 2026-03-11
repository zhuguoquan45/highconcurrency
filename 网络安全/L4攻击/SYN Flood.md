## SYN Flood 原理（SYN 洪泛攻击）

![Image](https://docs.aws.amazon.com/images/whitepapers/latest/aws-best-practices-ddos-resiliency/images/syn-three-way-handshake.png)

![Image](https://www.akamai.com/site/en/images/article/2022/syn-flood-attack-diagram.png)

![Image](https://www.cloudflare.com/img/learning/ddos/syn-flood-ddos-attack/syn-flood-attack-ddos-attack-diagram-1.png)

![Image](https://www.researchgate.net/publication/328512658/figure/fig4/AS%3A685626110275592%401540477411289/A-comparison-of-TCP-SYN-Flood-Attack-with-a-successful-Three-Way-Handshake.png)

**SYN Flood** 是一种典型的 **DoS / DDoS 网络攻击**，它利用 **TCP 三次握手（Three-Way Handshake）机制的弱点**，通过发送大量 **SYN 请求**但不完成连接，导致服务器资源被耗尽，从而无法为正常用户提供服务。

---

# 一、TCP 三次握手（正常连接）

TCP 建立连接需要 **三次握手**：

1️⃣ 客户端发送 **SYN**（请求建立连接）
2️⃣ 服务器回复 **SYN-ACK**（确认请求）
3️⃣ 客户端发送 **ACK**（确认连接）

流程：

```
Client  → SYN      → Server
Client  ← SYN-ACK  ← Server
Client  → ACK      → Server
Connection Established
```

完成后才会开始数据通信。

---

# 二、SYN Flood 攻击原理

攻击者利用 **服务器在第二步需要等待 ACK 的机制**。

攻击流程：

1️⃣ 攻击者向服务器发送大量 **SYN 请求**
2️⃣ 服务器为每个请求创建 **半连接（Half-Open Connection）**
3️⃣ 服务器回复 **SYN-ACK** 并等待客户端 ACK
4️⃣ 攻击者 **不发送 ACK**

结果：

* 服务器的 **连接队列（SYN Queue）被占满**
* 新用户无法建立连接

攻击流程示意：

```
Attacker → SYN → Server
Server   → SYN-ACK → Attacker
Attacker (no response)

Server keeps waiting...
```

当这种请求 **成千上万次**出现时：

* 服务器资源被耗尽
* 正常连接被拒绝

---

# 三、为什么 SYN Flood 有效

主要原因有三个：

### 1️⃣ 半连接会占用资源

服务器会为每个 SYN 请求分配：

* 内存
* 连接表记录
* 等待计时器

---

### 2️⃣ 服务器需要等待超时

如果没有 ACK：

服务器会等待 **几十秒**才释放连接。

---

### 3️⃣ 攻击成本低

攻击者只需要发送 **SYN 包**，流量很小，但服务器消耗很大。

---

# 四、常见 SYN Flood 攻击方式

### 1️⃣ Direct SYN Flood

攻击者直接发送 SYN。

```
Attacker → Target Server
```

---

### 2️⃣ Spoofed SYN Flood

使用 **伪造 IP 地址**。

```
Attacker → Fake IP → Server
```

服务器回复给假地址，连接永远不会完成。

---

### 3️⃣ Distributed SYN Flood（DDoS）

利用 **僵尸网络（Botnet）**。

```
Bot1
Bot2 → Target Server
Bot3
Bot4
```

这种攻击流量巨大。

---

# 五、攻击造成的影响

可能导致：

* 网站无法访问
* API 服务中断
* 服务器 CPU / 内存占满
* 网络设备崩溃

---

# 六、防御方法

### 1️⃣ SYN Cookies

服务器在收到 SYN 时 **不立即分配资源**，而是使用加密 Cookie。

Linux 示例：

```bash id="yn26o0"
sysctl -w net.ipv4.tcp_syncookies=1
```

---

### 2️⃣ 增大 SYN Queue

增加半连接队列容量。

---

### 3️⃣ 防火墙 / IDS

检测异常 SYN 流量。

---

### 4️⃣ DDoS 防护服务

例如：

* Cloudflare
* AWS Shield
* Akamai

---

✅ **总结**

SYN Flood 的核心原理：

```
大量 SYN 请求
       ↓
服务器创建半连接
       ↓
攻击者不发送 ACK
       ↓
连接队列被占满
       ↓
正常用户无法连接
```

---

如果你愿意，我还可以继续帮你整理 **网络安全常见 DoS 攻击对比表**（很多网络安全考试和面试都会考）：

* ICMP Flood
* SYN Flood
* UDP Flood
* Ping of Death
* Smurf Attack
* IP Fragment Attack 🔐
