# context 包

## 核心作用

context 用于在 goroutine 之间传递**取消信号、超时、截止时间和请求级别的值**，是 Go 高并发编程的基础设施。

---

## 四种 context 类型

```go
// 1. 根 context，永不取消
ctx := context.Background()

// 2. 手动取消
ctx, cancel := context.WithCancel(parent)
defer cancel() // 必须调用，否则泄漏

// 3. 超时（相对时间）
ctx, cancel := context.WithTimeout(parent, 3*time.Second)
defer cancel()

// 4. 截止时间（绝对时间）
ctx, cancel := context.WithDeadline(parent, time.Now().Add(3*time.Second))
defer cancel()

// 5. 携带值（仅用于请求级别数据，不要传业务参数）
ctx = context.WithValue(parent, key, value)
```

---

## 取消传播

context 是树形结构，父 context 取消会自动取消所有子 context。

```
Background
    └── WithCancel (A)
            ├── WithTimeout (B) ← A 取消时，B 也取消
            └── WithValue (C)   ← A 取消时，C 也取消
```

---

## 标准用法

### HTTP 请求链路控制
```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // 请求自带 context，客户端断开时自动取消

    result, err := queryDB(ctx)
    if err != nil {
        // ctx.Err() == context.Canceled 说明客户端已断开
        return
    }
}
```

### goroutine 生命周期控制
```go
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            // 收到取消信号，退出
            return
        default:
            // 正常工作
            doWork()
        }
    }
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
go worker(ctx)
```

### 数据库/HTTP 客户端超时
```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

// 数据库查询超时
row := db.QueryRowContext(ctx, "SELECT ...")

// HTTP 请求超时
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

---

## WithValue 使用规范

```go
// 用自定义类型作 key，避免冲突
type contextKey string
const requestIDKey contextKey = "requestID"

ctx = context.WithValue(ctx, requestIDKey, "req-123")
id := ctx.Value(requestIDKey).(string)
```

**注意：**
- 只存请求级别数据（requestID、traceID、userID）
- 不要存业务逻辑参数，用函数参数传递
- Value 查找是线性的，不要存大量数据

---

## 面试高频问题

**Q: 为什么要用 context 而不是全局变量？**
- context 是请求级别的，每个请求独立
- 全局变量并发不安全，且无法自动清理
- context 支持取消传播，全局变量做不到

**Q: context 取消后，goroutine 会立即停止吗？**
- 不会，取消只是发送信号（关闭 Done channel）
- goroutine 需要主动监听 `ctx.Done()` 并退出
- 不监听的 goroutine 不受影响（这是泄漏的常见原因）

**Q: WithTimeout 和 WithDeadline 的区别？**
- WithTimeout：相对时间，`time.Now().Add(d)`
- WithDeadline：绝对时间点
- 本质相同，WithTimeout 内部调用 WithDeadline

**Q: cancel 函数为什么必须调用？**
- 即使 context 已超时，不调用 cancel 会导致资源泄漏
- context 内部有 goroutine 监听定时器，cancel 负责清理

**Q: context 在微服务中如何传递？**
- gRPC：通过 metadata 传递，服务端用 `metadata.FromIncomingContext`
- HTTP：通过 Header 传递 traceID，服务端重建 context
- 链路追踪（OpenTelemetry）基于 context 传播 span
