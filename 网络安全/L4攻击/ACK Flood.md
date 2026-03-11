## ACK Flood 攻击原理（TCP ACK Flood）

![Image](https://www.cloudflare.com/img/learning/cdn/tls-ssl/tcp-handshake-diagram.png)

![Image](https://www.researchgate.net/publication/399956149/figure/fig2/AS%3A11431281875762137%401769015299561/ACK-flood-DDoS-attacks-schematic-diagram.png)

![Image](https://www.cloudflare.com/img/learning/ddos/syn-flood-ddos-attack/syn-flood-attack-ddos-attack-diagram-2.png)

![Image](https://www.researchgate.net/publication/320654932/figure/fig10/AS%3A629880433680386%401527186606914/The-TCP-SYN-flood-attack.png)

**ACK Flood** 是一种 **TCP 层 DoS / DDoS 攻击**，攻击者发送大量 **伪造的 ACK 包** 到目标服务器，利用 TCP 协议特点使服务器消耗大量资源，从而导致服务中断。

与 **SYN Flood** 不同，ACK Flood 攻击不需要建立连接，因为攻击包通常 **直接带 ACK 标志**，服务器仍然需要进行 **连接状态处理和包验证**。

---

# 一、ACK Flood 工作原理

1️⃣ 攻击者发送大量 **带 ACK 标志的 TCP 包** 给目标服务器。
2️⃣ 目标服务器收到 ACK 后，会尝试匹配 **已有的连接状态（TCP 状态表）**。
3️⃣ 如果 ACK 无效（不存在对应连接）：

* 服务器会生成 **RST 响应** 或尝试查找半连接
* 每个包都消耗 **CPU / 内存** 资源
  4️⃣ 攻击包数量巨大时：
* 服务器资源被耗尽
* 网络带宽占用
* 正常连接无法建立

---

# 二、攻击特点

| 特性      | 描述                       |
| ------- | ------------------------ |
| 不需要三次握手 | 包含 ACK 标志即可，不需要建立 TCP 连接 |
| 消耗 CPU  | 每个包都需要服务器验证 ACK 是否有效     |
| 高效率     | 小流量就能造成服务拒绝              |
| DDoS 方式 | 可以使用 Botnet 产生大规模攻击      |

---

# 三、与 SYN Flood 对比

| 特性   | SYN Flood          | ACK Flood             |
| ---- | ------------------ | --------------------- |
| 攻击方式 | 发大量 SYN 包，占半连接队列   | 发大量 ACK 包，占服务器处理能力    |
| 目标资源 | 半连接队列              | TCP 状态表 + CPU         |
| 攻击成本 | 低                  | 低                     |
| 防御方法 | SYN Cookies / 增大队列 | 防火墙 / TCP 报文过滤 / 流量清洗 |

---

# 四、防御方法

### 1️⃣ 防火墙过滤异常 ACK

例如 Linux `iptables`：

```bash
iptables -A INPUT -p tcp --tcp-flags ALL ACK -m limit --limit 1/s -j ACCEPT
```

### 2️⃣ 流量清洗中心

* 云防护（Cloudflare / AWS Shield）
* 基于 BGP 或 CDN 过滤恶意流量

### 3️⃣ TCP 堆栈优化

* 增加连接状态表容量
* 使用硬件防火墙限制异常 TCP 包

---

# 五、总结

**ACK Flood 攻击核心：**

```text
大量 TCP ACK 包
        ↓
服务器验证 ACK 状态
        ↓
CPU / 内存消耗
        ↓
正常连接被阻塞
```

> 与 SYN Flood 不同，ACK Flood 不占用半连接队列，而是直接消耗服务器 **TCP 状态表和处理能力**，适用于高带宽 / 高连接服务器攻击。

---

我可以帮你画一张 **“SYN Flood vs ACK Flood vs ICMP Flood 对比攻击流程图”**，把面试中经常问的 **三类网络层攻击原理**直观展示出来。你想让我画吗？
