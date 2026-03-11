# pprof 性能分析

## 什么是 pprof

Go 内置的性能分析工具，可以分析 CPU、内存、goroutine、锁竞争等性能瓶颈。

---

## 接入方式

### HTTP 方式（推荐生产环境）
```go
import _ "net/http/pprof"

go func() {
    http.ListenAndServe(":6060", nil)
}()
```

访问 `http://localhost:6060/debug/pprof/` 查看所有 profile。

### 代码方式
```go
import "runtime/pprof"

// CPU profile
f, _ := os.Create("cpu.prof")
pprof.StartCPUProfile(f)
defer pprof.StopCPUProfile()

// 内存 profile
f, _ := os.Create("mem.prof")
pprof.WriteHeapProfile(f)
```

---

## 常用 Profile 类型

| Profile | 说明 | 接口 |
|---------|------|------|
| cpu | CPU 使用热点 | /debug/pprof/profile?seconds=30 |
| heap | 堆内存分配 | /debug/pprof/heap |
| goroutine | 所有 goroutine 堆栈 | /debug/pprof/goroutine |
| block | goroutine 阻塞点 | /debug/pprof/block |
| mutex | 锁竞争 | /debug/pprof/mutex |
| allocs | 内存分配（含已释放） | /debug/pprof/allocs |

---

## 分析命令

```bash
# 采集 30 秒 CPU profile 并分析
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 分析本地 profile 文件
go tool pprof cpu.prof

# 常用交互命令
(pprof) top        # 显示 CPU 占用最高的函数
(pprof) top -cum   # 按累计时间排序
(pprof) list func  # 查看某函数的源码级分析
(pprof) web        # 生成火焰图（需安装 graphviz）

# 生成火焰图（推荐）
go tool pprof -http=:8080 cpu.prof
```

---

## 实战排查场景

### CPU 占用高
```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
(pprof) top 10
(pprof) list hotFunction
```
关注：正则匹配、JSON 序列化、频繁内存分配

### 内存泄漏
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
(pprof) top -cum
```
关注：goroutine 持有的大对象、全局 map 无限增长

### goroutine 泄漏
```bash
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```
关注：大量 goroutine 阻塞在同一位置

### 锁竞争
```go
// 需要先开启 mutex profile
runtime.SetMutexProfileFraction(1)
```
```bash
go tool pprof http://localhost:6060/debug/pprof/mutex
```

---

## benchmark + pprof 结合

```bash
# 运行 benchmark 并生成 CPU profile
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof

go tool pprof cpu.prof
```

---

## 面试高频问题

**Q: 如何排查 Go 服务 CPU 飙高？**
1. 接入 pprof HTTP 端点
2. 采集 30s CPU profile
3. `top` 找热点函数，`list` 看源码
4. 常见原因：正则、JSON、锁竞争、GC 频繁

**Q: 如何排查内存持续增长？**
1. 采集 heap profile（间隔采集对比）
2. 关注 inuse_space（当前占用）vs alloc_space（历史分配）
3. 常见原因：goroutine 泄漏、全局缓存无限增长、sync.Pool 误用

**Q: block profile 和 mutex profile 的区别？**
- block：goroutine 在 channel/select/sync 上的阻塞时间
- mutex：锁的竞争等待时间
- 两者都需要手动开启采样率
