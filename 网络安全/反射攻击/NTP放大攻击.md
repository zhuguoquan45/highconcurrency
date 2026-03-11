## NTP Amplification Attack 原理（NTP 放大攻击）

![Image](https://www.cloudflare.com/img/learning/ddos/ntp-amplification-ddos-attack/ntp-amplification-attack-ddos-attack-diagram-2.png)

![Image](https://www.akamai.com/site/en/images/article/2024/akamai-what-is-ntp-ddos.png)

![Image](https://cf-assets.www.cloudflare.com/zkvhlag99gkb/5AUH2Pw2eHluzo2Azh5WWJ/4093ce97e0bdc109f0e86d1847ce002a/illustration-amplification-attack-ph3.png)

**NTP Amplification Attack** 是一种典型的 **UDP 放大型 DDoS 攻击**，利用 **NTP（Network Time Protocol）服务器的 MONLIST 命令** 或其他查询请求放大流量，攻击目标服务器。

---

# 一、NTP 协议简介

* NTP 用于 **网络时间同步**
* 使用 **UDP 端口 123**
* 客户端向 NTP 服务器发送请求，服务器返回响应

正常 NTP 查询：

```
Client → NTP Request (UDP 123)
NTP Server → NTP Response (UDP 123)
```

特点：

* 请求包小
* 响应包可能大（尤其是 MONLIST 命令，可返回最近 600 个客户端信息）

---

# 二、NTP Amplification 攻击原理

攻击利用 NTP **放大和反射机制**：

1️⃣ 攻击者发送 **NTP 请求包**
2️⃣ **伪造源 IP 为目标服务器 IP**
3️⃣ 发送给 **开放 NTP 服务器**
4️⃣ NTP 服务器返回 **大流量响应** 到目标 IP

流程示意：

```text
Attacker → NTP Request (Source IP = Victim) → Open NTP Server
NTP Server → Large NTP Response → Victim
```

* 单个请求可能被放大 **10–200 倍**
* 如果使用 **成百上千台开放 NTP 服务器**，攻击流量可达到 **Tbps** 级别

---

# 三、攻击特点

| 特性         | 描述                   |
| ---------- | -------------------- |
| 协议         | UDP 123              |
| 无需建立连接     | UDP 无连接，易伪造源 IP      |
| 放大倍数大      | 小请求 → 大响应，放大可达 200 倍 |
| 高效率        | 少量流量可产生大规模攻击         |
| Reflection | 利用开放 NTP 服务器反射攻击目标   |

---

# 四、典型攻击命令

**MONLIST**：

* 返回 NTP 服务器最近与之通信的 **客户端列表**
* 响应数据可能很大
* 攻击者伪造源 IP，目标会收到大流量

示例：

```
Attacker → UDP 123, MONLIST → Open NTP Server
Open NTP Server → Large UDP response → Victim
```

> 现代 NTP 版本已禁用 MONLIST，但仍有部分旧服务器可被利用

---

# 五、防御方法

### 1️⃣ 禁用 MONLIST

* 更新 NTP 服务器
* 禁用易被放大的命令

### 2️⃣ 限制开放访问

* 仅允许可信 IP 查询 NTP

### 3️⃣ BCP38 防止 IP 伪造

* ISP 阻止源地址伪造，减少反射攻击可能性

### 4️⃣ DDoS 防护

* 云防护、流量清洗中心、CDN 等

---

# 六、放大倍数示意

| 请求大小     | 响应大小       | 放大倍数 |
| -------- | ---------- | ---- |
| 60 bytes | 4000 bytes | ~66× |
| 50 bytes | 3200 bytes | ~64× |

---

# 七、总结

**NTP Amplification 核心原理**：

```text
小 NTP 请求
       ↓
伪造源 IP = 目标
       ↓
开放 NTP 服务器响应
       ↓
大流量打向目标
       ↓
形成 DDoS 攻击
```

> 核心就是 **Reflection（反射） + Amplification（放大）**，与 DNS Amplification 类似，但使用 NTP 协议。

---

我可以帮你做一张 **“DNS vs NTP vs UDP vs ICMP Amplification 攻击流程对比图”**，把面试常考的几类放大攻击可视化，理解更直观。你希望我画吗？
