## SSDP Amplification Attack 原理（SSDP 放大攻击）

![Image](https://www.cloudflare.com/img/learning/ddos/ssdp-ddos-attack/ping-of-death.png)

![Image](https://www.researchgate.net/publication/349457868/figure/fig5/AS%3A1008221997260800%401617390264203/SSDP-reflection-attacks-procedures.png)

![Image](https://content2.stormwall.network/or/articles/picture-ssdp-2_w950.jpg)

**SSDP Amplification** 是一种基于 **UPnP（Universal Plug and Play）协议**的 **UDP 放大型 DDoS 攻击**。攻击者利用 **SSDP 服务（UDP 1900端口）的反射和放大特性**，对目标服务器发起大流量攻击。

---

# 一、SSDP 协议简介

* SSDP 是 **UPnP 的一部分**
* 用于 **设备发现和服务发现**
* 工作在 **UDP 端口 1900**
* 客户端发送 **M-SEARCH 请求**，设备返回响应列表

正常流程：

```text
Client → M-SEARCH → UPnP Device
UPnP Device → Response → Client
```

特点：

* 请求包很小
* 响应可能大很多（包含设备信息列表）

---

# 二、SSDP Amplification 攻击原理

攻击者利用 **UDP 无连接 + 放大响应 + IP 伪造**：

1️⃣ 攻击者向 **开放的 SSDP 设备**发送 M-SEARCH 请求
2️⃣ **伪造源 IP 为目标服务器 IP**
3️⃣ SSDP 设备返回 **大量响应数据** 到目标 IP

流程示意：

```text
Attacker → UDP 1900, M-SEARCH (Source IP = Victim) → Open SSDP Device
Open SSDP Device → Large UDP Response → Victim
```

特点：

* 单个请求放大倍数可达 **30–50 倍**
* 利用成百上千设备可产生 **大规模 DDoS 流量**

---

# 三、攻击特点

| 特性         | 描述                     |
| ---------- | ---------------------- |
| 协议         | UDP 1900（SSDP / UPnP）  |
| 放大倍数       | 小请求 → 大响应（可放大 30–50 倍） |
| Reflection | 利用开放 SSDP 设备反射攻击       |
| 无需建立连接     | UDP 无连接，易伪造源 IP        |
| 高效率        | 少量攻击流量产生巨大带宽冲击         |

---

# 四、防御方法

### 1️⃣ 禁止开放 SSDP

* 在公网关闭 UPnP / SSDP 服务
* 屏蔽 UDP 1900 端口

### 2️⃣ 限制响应

* UPnP 设备限制 M-SEARCH 响应频率
* 配置防火墙限制异常请求

### 3️⃣ BCP38 防止 IP 伪造

* ISP 阻止源地址伪造，防止反射攻击

### 4️⃣ DDoS 清洗中心

* 云防护 / 流量清洗
* CDN 屏蔽异常流量

---

# 五、攻击效果

* 目标带宽被占满
* 网络延迟高、服务不可用
* 适合 **大规模反射放大攻击**（例如 2014 年 Mirai 僵尸网络攻击）

---

# 六、总结

**SSDP Amplification 攻击核心原理：**

```text
小 M-SEARCH 请求
       ↓
伪造源 IP = 目标
       ↓
开放 SSDP 设备响应大量数据
       ↓
形成 DDoS 攻击
```

> 核心就是 **Reflection（反射）+ Amplification（放大）**，与 **DNS / NTP / Memcached 放大攻击** 类似，但攻击媒介是 **SSDP / UPnP 设备**。

---

我可以帮你做一张 **“常见 Amplification 放大攻击对比图”**，把 **DNS / NTP / SSDP / Memcached** 这些放大攻击直观展示，包括**请求大小、响应大小、放大倍数和攻击原理流程**，非常适合面试快速讲解。你希望我画吗？
