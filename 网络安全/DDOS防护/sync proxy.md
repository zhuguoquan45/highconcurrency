## TCP SYN Proxy 实现原理

![Image](https://www.researchgate.net/publication/224315331/figure/fig8/AS%3A302732847140868%401449188543186/The-SYN-Proxy-method-against-TCP-SYN-flood.png)

![Image](https://www.researchgate.net/publication/336872866/figure/fig3/AS%3A956296866050053%401605010348138/Protection-against-SYN-flooding-using-SYN-Proxy-7.jpg)

![Image](https://iximiuz.com/laymans-iptables-101/iptables-stages-white.png)

![Image](https://upload.wikimedia.org/wikipedia/commons/3/37/Netfilter-packet-flow.svg)

**SYN Proxy** 是一种针对 **SYN Flood 攻击** 的 **内核级防护机制**，通过 **代理三次握手** 来保护服务器半连接队列不被耗尽。它常用于 **Linux iptables/netfilter** 或专业防火墙。

---

# 一、SYN Proxy 原理

SYN Proxy 的核心思想：

1️⃣ **代理客户端与服务器之间的三次握手**
2️⃣ 在服务器真正分配资源前，先 **验证客户端是否有效**
3️⃣ 仅对合法客户端创建实际 TCP 连接

流程：

```text id="xv5hr3"
Client → SYN → SYN Proxy
SYN Proxy → SYN → Real Server
SYN Proxy ← SYN-ACK ← Real Server
SYN Proxy → ACK → Real Server
SYN Proxy ← ACK ← Client
Connection Established
```

解释：

* **客户端只与 SYN Proxy 完成三次握手**
* **服务器只看到 SYN Proxy 发起的连接**
* 可以阻止大量伪造 SYN 包直接占用服务器半连接队列

---

# 二、工作机制

### 1️⃣ 初始 SYN 验证

* SYN Proxy 接收客户端 SYN
* 不立即转发到服务器
* 使用 **SYN Cookie** 或 **序列号计算** 验证客户端 ACK

### 2️⃣ 完成握手再转发

* 客户端返回 ACK → SYN Proxy 验证
* 代理与服务器建立真正连接
* 服务器仅处理 **合法连接**

### 3️⃣ 半连接保护

* 防止服务器半连接队列被 **伪造 SYN 包占满**
* 只为验证合法客户端创建半连接

---

# 三、Linux iptables SYNPROXY 示例

```bash id="s0n72f"
# 内核选项启用
sysctl -w net.ipv4.tcp_syncookies=1

# iptables 配置 SYNPROXY
iptables -t raw -A PREROUTING -p tcp --syn -j CT --notrack
iptables -A INPUT -p tcp -m state --state INVALID,UNTRACKED \
    -j SYNPROXY --sack-perm --timestamp --wscale 7 --mss 1460
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
```

解释：

* **`--notrack`**: 临时不在连接追踪表记录
* **SYNPROXY 参数**：

  * `--sack-perm`：支持 SACK
  * `--timestamp`：支持时间戳
  * `--wscale`：窗口缩放
  * `--mss`：最大报文长度

SYN Proxy 可以在 **内核级别快速处理 SYN 包**，极大减少服务器资源消耗。

---

# 四、优缺点

| 优点           | 描述                             |
| ------------ | ------------------------------ |
| 防御 SYN Flood | 代理三次握手，保护服务器半连接队列              |
| 高效率          | 内核实现，延迟低                       |
| 支持 TCP 选项    | SACK、Timestamp、Window Scale 可选 |

| 缺点             | 描述                         |
| -------------- | -------------------------- |
| 配置复杂           | 需要 iptables / netfilter 配合 |
| 部分 TCP 特性限制    | 仍可能影响某些高级 TCP 功能           |
| 对大规模攻击，需要高性能网络 | 防火墙自身需承载大量连接               |

---

# 五、总结

**SYN Proxy 核心原理**：

```text id="q0c2zi"
客户端 SYN → SYN Proxy
          ↓ 验证合法性（SYN Cookie）
客户端 ACK → SYN Proxy → 建立真正 TCP 连接到服务器
          ↓
服务器处理合法连接，半连接队列安全
```

> 核心是 **代理握手 + 延迟资源分配 + SYN Cookie 验证**，有效防止 SYN Flood 攻击占用服务器资源。

---

我可以帮你画一张 **“SYN Proxy 与普通三次握手对比图 + 防护流程图”**，直观展示它如何保护半连接队列，非常适合面试讲解。你希望我画吗？
