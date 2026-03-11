## DNS Amplification Attack 原理（DNS 放大攻击）

![Image](https://cf-assets.www.cloudflare.com/slt3lc6tev37/2JmKP07Mi6jYbACILN84VI/9a91d91ecc1f414aa89ae001dbfce393/Learning_Center_DDoS_Diagrams_clean.png)

![Image](https://www.cdnetworks.com/wos/static-resource/d42044fefe6b4e90947f25e8c9ad69ff/DNS-Resolver-Attack-security-blog1.png?t=1742205805071)

![Image](https://cdn.prod.website-files.com/5ff66329429d880392f6cba2/67b431ba6f46417ac2df99c9_61a08ec5191af44ea3b67fb7_DNS%2520Amplification%2520Attacks%2520Works.png)

**DNS Amplification Attack** 是一种常见的 **DDoS（分布式拒绝服务）攻击**，利用 **DNS 协议的放大特性（Amplification）和反射机制（Reflection）**，通过少量请求产生大量响应流量攻击目标服务器。

---

# 一、DNS 的正常工作流程

DNS（Domain Name System）用于 **域名解析**。

正常流程：

1️⃣ 客户端向 DNS 服务器发送查询请求
2️⃣ DNS 服务器返回解析结果

示例：

```text
Client → DNS Query (example.com)
DNS Server → DNS Response (IP Address)
```

通常请求数据包 **很小**，响应数据包 **更大**。

---

# 二、DNS Amplification 攻击原理

攻击者利用 **UDP 协议 + IP 伪造 + DNS 放大**。

攻击步骤：

1️⃣ 攻击者发送 **DNS 查询请求**
2️⃣ **伪造源 IP 为目标服务器 IP**
3️⃣ 请求发送给 **开放 DNS 解析器（Open Resolver）**
4️⃣ DNS 服务器把 **大量响应数据** 发给目标服务器

流程：

```text
Attacker → DNS Query (Fake Source = Victim)
                ↓
         Open DNS Resolver
                ↓
DNS Response (Large Data) → Victim
```

结果：

* 目标服务器收到 **大量 DNS 响应流量**
* 带宽被占满
* 服务无法访问

---

# 三、为什么叫 “Amplification（放大）”

原因是 **响应数据比请求大很多**。

示例：

| 数据类型         | 大小         |
| ------------ | ---------- |
| DNS Query    | 60 bytes   |
| DNS Response | 3000 bytes |

放大倍数：

```
3000 / 60 = 50 倍
```

如果攻击者发送：

```
1 Gbps 请求
```

目标可能收到：

```
50 Gbps 攻击流量
```

---

# 四、攻击的关键技术

### 1️⃣ IP Spoofing（IP 伪造）

攻击者伪造源 IP，使 DNS 响应发送给目标。

---

### 2️⃣ UDP 协议

UDP 是 **无连接协议**：

* 不验证源地址
* 可以轻易伪造

---

### 3️⃣ Open DNS Resolver

攻击者利用 **开放 DNS 服务器**作为反射器。

---

# 五、常见放大攻击协议

DNS 只是其中一种，类似攻击还有：

| 协议        | 放大攻击                    |
| --------- | ----------------------- |
| DNS       | DNS Amplification       |
| NTP       | NTP Amplification       |
| SSDP      | SSDP Amplification      |
| Memcached | Memcached Amplification |
| CLDAP     | LDAP Amplification      |

---

# 六、DNS Amplification 的影响

可能造成：

* 大规模 **DDoS 攻击**
* 网络带宽被占满
* 网站或 API 无法访问
* ISP 网络拥塞

历史上很多 **Tbps 级攻击** 都使用 DNS 放大。

---

# 七、防御方法

### 1️⃣ 关闭 Open Resolver

DNS 服务器只允许内部网络查询。

---

### 2️⃣ Response Rate Limiting (RRL)

限制 DNS 响应速率。

---

### 3️⃣ BCP38（防止 IP Spoofing）

ISP 阻止伪造 IP 地址。

---

### 4️⃣ DDoS 防护

使用：

* Cloudflare
* AWS Shield
* Akamai

---

# 八、总结

DNS Amplification 攻击核心原理：

```text
小 DNS 请求
      ↓
伪造源 IP
      ↓
开放 DNS 服务器响应
      ↓
大流量返回目标
      ↓
形成 DDoS
```

本质是：

```
Reflection（反射） + Amplification（放大）
```

---

如果你需要，我可以再帮你整理一张 **“常见 10 种 DDoS 攻击原理对比表”**（ICMP、SYN、UDP、DNS、NTP、Smurf 等），这在 **网络安全考试和面试**里非常常见。
