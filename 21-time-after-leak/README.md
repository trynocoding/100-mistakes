# time.After 内存泄漏

## 概念

`time.After()` 是 Go 中常用的定时器函数，但在循环中使用它会导致严重的内存泄漏问题。每次调用 `time.After()` 都会创建一个新的定时器，该定时器在到期之前无法被垃圾回收（GC）。如果在循环中频繁调用 `time.After()`，会不断创建新的定时器对象，导致内存持续增长。

## 问题根源

`time.After()` 返回一个 `<-chan time.Time`，每次调用都会：

1. 创建一个新的 `*time.Timer` 对象
2. 启动一个新的 goroutine 来处理定时器
3. 返回一个通道，直到时间到期才会被关闭

关键问题：**定时器在通道被读取之前不会被 GC 回收**。即使循环继续执行，旧定时器占用的内存也不会被释放。

## 错误示例

### 示例 1：循环内使用 time.After

```go
func processWithDelay(items []string) {
    for _, item := range items {
        select {
        case <-time.After(time.Second): // 错误：每次迭代都创建新定时器
            fmt.Println("处理:", item)
        case <-ctx.Done():
            return
        }
    }
}
```

这段代码的问题在于，每次迭代都会创建一个新的定时器，导致：
- 内存持续增长
- 如果处理大量数据，内存可能爆炸

### 示例 2：心跳检测中的泄漏

```go
func heartbeatCheck(ctx context.Context) {
    for {
        select {
        case <-time.After(5 * time.Second): // 错误：每次循环都创建新定时器
            if err := checkHealth(); err != nil {
                log.Printf("健康检查失败: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

## 正确写法

### 方法 1：使用 time.NewTimer() 并定期 Reset

```go
func processWithDelayFixed(items []string) {
    timer := time.NewTimer(time.Second)
    defer timer.Stop() // 确保定时器被正确关闭

    for _, item := range items {
        // 重置定时器，复用同一个定时器对象
        if !timer.Stop() {
            <-timer.C // 必须消耗已触发的信号
        }
        timer.Reset(time.Second)

        select {
        case <-timer.C:
            fmt.Println("处理:", item)
        case <-ctx.Done():
            return
        }
    }
}
```

### 方法 2：使用 time.Ticker（适用于周期性任务）

```go
func heartbeatCheckFixed(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C: // 复用同一个 ticker
            if err := checkHealth(); err != nil {
                log.Printf("健康检查失败: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

## 内存泄漏对比

| 方法 | 定时器数量 | GC 压力 | 适用场景 |
|------|-----------|---------|---------|
| `time.After` 在循环中 | 每次迭代新建 | 高 | 不推荐 |
| `time.NewTimer` + `Reset` | 1个 | 低 | 单次延迟后可重置 |
| `time.Ticker` | 1个 | 低 | 周期性任务 |

## 常见误区

1. **误以为 `time.After` 会自动清理**：实际上，定时器对象在通道被读取之前会一直保留在内存中

2. **在 select 中使用 `time.After` 是安全的**：这取决于使用场景。如果是循环中使用，仍然会有泄漏问题

3. **`timer.Stop()` 就能释放内存**：对于已经创建的定时器，Stop 只是停止了定时器，但通道和数据结构仍然存在直到 GC

## 最佳实践

1. **在循环中使用定时器时**：始终使用 `time.NewTimer()` + `Reset()` 或 `time.Ticker`
2. **对于周期性任务**：优先使用 `time.Ticker`
3. **对于需要取消的定时器**：记得调用 `Stop()` 或使用 `context.WithTimeout`
4. **处理 Reset 后的定时器**：如果定时器已经触发，先读取通道再 Reset

## 验证方法

### 使用 pprof 检查内存

```bash
# 启动服务
go run -cpuprofile=cpu.prof -memprofile=mem.prof main.go

# 分析内存
go tool pprof -http=:8080 mem.prof
```

### 观察定时器数量

```go
import "runtime/debug"

func printTimerStats() {
    debug.FreeOSMemory() // 尝试释放内存
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    fmt.Printf("定时器/对象数量: %d\n", m.Mallocs)
}
```

## 参考

- [Go time package documentation](https://pkg.go.dev/time)
- [Runtime timer implementation](https://github.com/golang/go/blob/master/src/runtime/time.go)
- [Go Wiki: Timer and Ticker](https://github.com/golang/go/wiki/CodeReviewComments#time-to-day)
