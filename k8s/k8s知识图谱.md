下面给你一份 **Kubernetes + Istio 知识图谱（结构化版）**。
我会用“分层 + 关联关系”的方式展示，方便你建立整体认知（尤其适合你现在在做服务网格和访问控制的学习）。

---

# 一、Kubernetes 知识图谱

## 1️⃣ 控制面（Control Plane）

![Image](https://kubernetes.io/images/docs/components-of-kubernetes.svg)

![Image](https://devopscube.com/content/images/2025/03/02-k8s-architecture-api-1.gif)

![Image](https://cdn.prod.website-files.com/681e366f54a6e3ce87159ca4/68f62d227531fe67cb7effd4_how-to-monitor-kube-controller-manager-diagram.webp)

![Image](https://kubernetes.feisky.xyz/~gitbook/image?dpr=3\&quality=100\&sign=dc818413\&sv=2\&url=https%3A%2F%2F4011040894-files.gitbook.io%2F~%2Ffiles%2Fv0%2Fb%2Fgitbook-x-prod.appspot.com%2Fo%2Fspaces%252F-LDAOok5ngY4pc1lEDes-1972196547%252Fuploads%252Fgit-blob-1ea9ef0f0f29c3247eb79d8e40b1890fb29b89da%252Fpost-ccm-arch%2520%284%29.png%3Falt%3Dmedia\&width=768)

### 核心组件

* **kube-apiserver**

  * 集群唯一入口
  * 所有对象 CRUD
  * RBAC / Admission / CRD

* **etcd**

  * 存储所有集群状态
  * 强一致性 KV

* **kube-scheduler**

  * Pod → Node 绑定
  * 资源评估
  * 亲和性/污点容忍

* **kube-controller-manager**

  * Deployment Controller
  * Node Controller
  * Endpoint Controller
  * Job Controller

---

## 2️⃣ 工作节点（Node）

![Image](https://d33wubrfki0l68.cloudfront.net/2475489eaf20163ec0f54ddc1d92aa8d4c87c96b/e7c81/images/docs/components-of-kubernetes.svg)

![Image](https://kubernetes.io/images/docs/kubernetes-cluster-architecture.svg)

![Image](https://kubernetes.io/images/docs/services-iptables-overview.svg)

![Image](https://cdn.sanity.io/images/xinsvxfu/production/d7e538715d25eddc181230273506aa9e58bd62bf-1600x973.webp)

### 核心组件

* **kubelet**

  * 与 API Server 通信
  * 管理 Pod 生命周期

* **kube-proxy**

  * Service 转发
  * iptables / ipvs

* Container Runtime

  * containerd / cri-o
  * 运行容器

---

## 3️⃣ 核心资源对象

### Workload

* Pod
* Deployment
* StatefulSet
* DaemonSet
* Job / CronJob

### 网络

* Service

  * ClusterIP
  * NodePort
  * LoadBalancer
  * ExternalName
* Ingress
* Gateway API（新标准）
* EndpointSlice
* CNI 插件（Calico / Cilium）

### 存储

* PV / PVC
* StorageClass
* CSI

### 安全

* RBAC
* ServiceAccount
* NetworkPolicy
* Pod Security Admission

---

# 二、Istio 知识图谱

Istio 是运行在 Kubernetes 之上的服务网格。

---

## 1️⃣ 架构总览

![Image](https://istio.io/latest/docs/ops/deployment/architecture/arch.svg)

![Image](https://iximiuz.com/service-proxy-pod-sidecar-oh-my/80-demo-sidecar.png)

![Image](https://d2908q01vomqb2.cloudfront.net/1b6453892473a467d07372d45eb05abc2031647a/2018/06/14/envoy-ecs.png)

### 两大平面

### 🔵 Data Plane（数据平面）

* Envoy Sidecar
* 每个 Pod 注入代理
* 所有流量都经过代理

能力：

* mTLS
* 熔断
* 重试
* 超时
* 负载均衡
* 流量镜像

---

### 🟣 Control Plane（控制平面）

* Istiod

  * Pilot（流量规则）
  * Citadel（证书）
  * Galley（已合并）

---

## 2️⃣ Istio 核心 CRD

### 流量治理

* VirtualService
* DestinationRule
* Gateway
* ServiceEntry

### 安全

* PeerAuthentication
* RequestAuthentication
* AuthorizationPolicy

### 可观测

* Telemetry
* EnvoyFilter

---

# 三、K8s 与 Istio 的关系图谱

## 1️⃣ 依赖关系

Kubernetes 提供：

* Pod
* Service
* DNS
* CNI 网络
* ServiceAccount
* CRD 机制

Istio 使用：

* CRD 扩展资源
* MutatingWebhook 注入 Sidecar
* ServiceAccount 做身份认证
* Service 作为流量目标

---

## 2️⃣ 流量路径对比

### 🔹 无 Istio

Client → Service → kube-proxy → Pod

### 🔹 有 Istio

Client
→ Service
→ Sidecar (Envoy)
→ mTLS 校验
→ 路由规则
→ 目标 Sidecar
→ 目标 Pod

---

# 四、完整知识图谱结构（层级表达）

```
云原生
 ├── Kubernetes
 │    ├── 控制面
 │    ├── 节点
 │    ├── 网络模型
 │    ├── 调度模型
 │    ├── 存储模型
 │    └── 安全模型
 │
 └── Istio
      ├── 数据平面（Envoy）
      ├── 控制平面（Istiod）
      ├── 流量治理
      ├── 安全（mTLS / RBAC）
      ├── 可观测
      └── 多集群
```

---

# 五、学习路径图（结合你目前问题）

你最近主要在研究：

* ServiceAccount
* 访问控制
* Gateway
* Service Mesh

建议路线：

### 第一阶段（K8s基础）

* Pod 生命周期
* Service 原理
* kube-proxy
* CNI 网络模型
* RBAC

### 第二阶段（安全模型）

* ServiceAccount
* Token
* RBAC
* NetworkPolicy

### 第三阶段（Istio 核心）

* Sidecar 原理
* VirtualService
* DestinationRule
* mTLS
* AuthorizationPolicy

### 第四阶段（生产级）

* 灰度发布
* A/B 测试
* 流量镜像
* 多集群
* 性能调优

---

# 六、终极关系一句话总结

> Kubernetes 解决的是 **“应用怎么跑”**
> Istio 解决的是 **“应用之间怎么通信”**

---

如果你愿意，我可以：

* 给你画一张“从 0 到生产级 Service Mesh 的完整进阶图”
* 或者做一个“面试级 Kubernetes + Istio 知识树”
* 或者专门做一个“访问控制（ServiceAccount + RBAC + Istio）全链路图”

你想往架构深度方向，还是面试梳理方向走？
