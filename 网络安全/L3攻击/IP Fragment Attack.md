## IP Fragment Attack 原理（IP 分片攻击）

![Image](https://cdn.prod.website-files.com/5ff66329429d880392f6cba2/6784e2d771f7fff06da3e458_612e282d58e1a41603c54525_tcp%2520based%2520Fragmentation%2520Attack.png)

![Image](https://cdn.prod.website-files.com/5ff66329429d880392f6cba2/612e27f35129788308df5de8_IP%20Fragmentation%20Attack%20Preview.png)

![Image](https://cdn.prod.website-files.com/5ff66329429d880392f6cba2/6710dfc159434f3eefd5babd_627b73886c890edf60373f02_Teardrop%2520attack%2520in%2520action.jpeg)

![Image](https://www.researchgate.net/publication/340929743/figure/fig16/AS%3A884417937539072%401587873076060/TCP-teardrop-fragmentation-attack-packet-scenario.png)

**IP Fragment Attack（IP 分片攻击）** 是一种利用 **IP 数据包分片（Fragmentation）机制的漏洞** 来攻击目标系统的方式。攻击者通过发送 **异常或恶意构造的 IP 分片数据包**，使目标主机在 **重组（Reassembly）数据包时出现错误、资源耗尽或系统崩溃**。

---

# 一、IP 分片机制（正常情况）

在 IP 网络中，如果一个数据包 **超过网络 MTU（Maximum Transmission Unit）**，路由器会把它 **拆分成多个小包**，这就是 **IP Fragmentation**。

示例：

原始 IP 包：

```
Original Packet (4000 bytes)
```

分片后：

```
Fragment 1 → 1500 bytes
Fragment 2 → 1500 bytes
Fragment 3 → 1000 bytes
```

这些分片会包含：

* **Identification**：用于标识属于同一个数据包
* **Fragment Offset**：分片位置
* **More Fragments (MF) flag**：是否还有后续分片

目标主机收到后会 **按照 offset 重组原始数据包**。

---

# 二、IP Fragment Attack 的核心原理

攻击者利用 **分片重组机制的弱点**。

攻击流程：

1️⃣ 攻击者发送 **大量分片数据包**
2️⃣ 这些分片可能：

* **重叠（overlapping fragments）**
* **顺序错误**
* **缺失某些分片**

3️⃣ 目标系统尝试 **重组这些异常分片**
4️⃣ 导致：

* 内存消耗
* CPU 占用
* 系统错误或崩溃

---

# 三、典型的 IP Fragment Attack 类型

## 1️⃣ Teardrop Attack

利用 **重叠分片（Overlapping Fragment）**。

示例：

```
Fragment A: offset 0 – 1000
Fragment B: offset 800 – 1500
```

出现 **数据重叠区域**：

```
800–1000
```

旧系统在处理时会：

* 发生错误
* 系统崩溃
* 蓝屏（Windows）

---

## 2️⃣ Fragment Flood

发送 **大量分片包**。

目标需要：

* 缓存分片
* 等待重组

结果：

* 内存被耗尽
* 服务拒绝（DoS）

---

## 3️⃣ Tiny Fragment Attack

把 TCP/UDP 头 **拆成非常小的片段**：

```
Fragment 1 → TCP Header Part
Fragment 2 → Rest of Packet
```

这样可以：

* 绕过防火墙
* 绕过 IDS 检测

---

# 四、为什么 IP Fragment Attack 有效

原因：

### 1️⃣ 系统需要缓存分片

系统必须保存分片直到重组完成。

---

### 2️⃣ 重组需要资源

需要：

* CPU
* 内存
* 网络缓冲区

---

### 3️⃣ 协议早期设计问题

旧系统在处理异常分片时：

* 没有完整验证
* 容易崩溃

---

# 五、攻击效果

可能造成：

* **系统崩溃**
* **网络设备卡死**
* **防火墙绕过**
* **DoS / DDoS**

---

# 六、防御方法

### 1️⃣ 丢弃异常分片

检测：

* 重叠分片
* 非法 offset

---

### 2️⃣ 限制分片缓存

减少 fragment buffer。

---

### 3️⃣ 使用现代防火墙 / IDS

例如：

* Snort
* Suricata
* Cisco ASA

---

### 4️⃣ 操作系统补丁

现代系统（Linux / Windows / BSD）已经修复大部分漏洞。

---

✅ **总结**

IP Fragment Attack 利用的核心：

```
IP 分片机制
      ↓
发送异常分片
      ↓
目标系统重组错误
      ↓
资源耗尽 / 系统崩溃
```

---

如果你感兴趣，我还可以继续讲更深入的网络安全知识，例如：

* **IP Fragment Attack vs Ping of Death 区别**
* **Wireshark 如何识别 Fragment Attack**
* **MTU、Fragment Offset 详细结构**
* **真实 DDoS 攻击案例分析** 🔍
