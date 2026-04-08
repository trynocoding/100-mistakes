# Goroutine 生命周期管理

## 目录

- [Goroutine 轻量但不是无限的](#goroutine-轻量但不是无限的)
- [确保 channel 关闭或使用 context 取消](#确保-channel-关闭或使用-context-取消)
- [主进程退出可能导致 goroutine 任务未完成](#主进程-退出可能导致-goroutine-任务未完成)
- [正确管理 goroutine 生命周期的模式](#正确管理-goroutine-生命周期的模式)
- [代码示例](#代码示例)

---

## Goroutine 轻量但不是无限的

Goroutine 是 Go 并发模型的基础，它比操作系统线程轻量得多，但这并不意味着我们可以随意创建无限数量的 goroutine。

### 内存开销

- 创建一个 goroutine 初始栈大小约为 2KB（相比线程的 1MB+）
- 栈会根据需要增长，但每个 goroutine 都有内存开销
- 无限创建 goroutine 会导致内存耗尽（OOM）

### 示例：无限创建 goroutine 的问题

```go
// BAD PRACTICE - 演示不推荐的做法
func processAll(items []string) {
    for _, item := range items {
        // 每个 item 都启动一个 goroutine
        // 如果 items 有 100万个元素，就会创建 100万个 goroutine
        go process(item)
    }
}
```

### 资源限制

- 文件描述符限制：每个 goroutine 可能需要打开网络连接
- 操作系统调度限制：大量 goroutine 会增加调度开销
- 正确的做法是使用 worker pool 限制并发数量

---

## 确保 channel 关闭或使用 context 取消

Goroutine 创建后需要有一种方式来通知它停止工作。主要有两种模式：

### 1. 通过 channel 关闭通知

使用一个专门的信号 channel 来通知 goroutine 退出。

```go
// 使用 done channel 通知退出
func worker(done chan bool) {
    for {
        select {
        case <-done:
            // 收到退出信号
            fmt.Println("Worker: 收到退出信号")
            done <- true // 确认退出
            return
        default:
            // 正常工作
        }
    }
}
```

### 2. 通过 context 取消

使用 `context.Context` 进行取消操作，这是最推荐的方式。

```go
func workerWithContext(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // 返回取消原因
        default:
            // 正常工作
        }
    }
}
```

### 为什么不能只依赖 channel 关闭？

- channel 关闭后，所有阻塞在该 channel 上的接收操作都会立即返回零值
- 如果有多个 goroutine 监听同一个 channel，无法区分是谁收到的信号
- context 可以携带取消原因，并且支持链式取消（父 context 取消，子 context 也会取消）

---

## 主进程退出可能导致 goroutine 任务未完成

这是新手最容易犯的错误：主函数结束后，goroutine 可能还没执行完。

### 问题演示

```go
// BAD PRACTICE - 演示不推荐的做法
func main() {
    go func() {
        // 模拟耗时任务
        time.Sleep(2 * time.Second)
        fmt.Println("后台任务完成")
    }()

    fmt.Println("主函数退出")
    // goroutine 可能还没执行完，进程就退出了
}
```

**输出**：可能只打印"主函数退出"，后台任务的打印永远不会出现。

### 原因

Go 程序的退出条件是：main 函数返回或者调用 `os.Exit()`。**不会等待** 所有 goroutine 执行完毕。

### 解决方案

1. 使用 `sync.WaitGroup` 等待所有 goroutine 完成
2. 使用 channel 同步
3. 使用 context + 取消信号

---

## 正确管理 goroutine 生命周期的模式

### 模式一：sync.WaitGroup 等待

```go
func main() {
    var wg sync.WaitGroup

    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            fmt.Printf("Goroutine %d 开始\n", id)
            time.Sleep(time.Second)
            fmt.Printf("Goroutine %d 完成\n", id)
        }(i)
    }

    fmt.Println("等待 goroutine 完成...")
    wg.Wait() // 阻塞直到计数器归零
    fmt.Println("所有 goroutine 已完成")
}
```

### 模式二：Context 取消 + Done Channel

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel() // 确保资源清理

    done := make(chan error, 1)

    go func() {
        done <- fetchUserData(ctx)
    }()

    select {
    case err := <-done:
        if err != nil {
            fmt.Println("任务失败:", err)
        }
    case <-time.After(3 * time.Second):
        cancel()
        fmt.Println("任务超时，已取消")
    }
}
```

### 模式三：Worker Pool + Context

```go
func workerPool(ctx context.Context, jobs <-chan int, results chan<- int) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("Worker: 收到取消信号，退出")
            return
        case job, ok := <-jobs:
            if !ok {
                // jobs channel 已关闭
                fmt.Println("Worker: jobs channel 已关闭，退出")
                return
            }
            // 处理任务
            results <- job * 2
        }
    }
}
```

### 模式四：多个 goroutine 共享同一个 Context

```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // 启动多个 worker，它们都会响应同一个 context
    var wg sync.WaitGroup
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go worker(ctx, i, &wg)
    }

    wg.Wait()
    fmt.Println("所有 worker 已完成或取消")
}

func worker(ctx context.Context, id int, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Worker %d: 取消原因: %v\n", id, ctx.Err())
            return
        case <-time.After(500 * time.Millisecond):
            fmt.Printf("Worker %d: 执行中...\n", id)
        }
    }
}
```

---

## 代码示例

请参见 [code.go](./code.go) 文件，包含以下示例：

1. **Goroutine 泄漏问题** - 展示未正确管理的 goroutine 如何导致资源泄漏
2. **Context 取消模式** - 展示如何使用 context 正确取消 goroutine
3. **Channel 关闭模式** - 展示如何使用 channel 信号通知 goroutine 退出
4. **Worker Pool 模式** - 展示如何限制并发数量并正确管理生命周期
5. **WaitGroup 等待模式** - 展示如何使用 WaitGroup 等待多个 goroutine 完成

### 运行代码

```bash
cd 16-goroutine-lifecycle
go run code.go
```

---

## 总结

### 常见错误

1. **无限创建 goroutine**：没有使用 worker pool 限制并发
2. **忘记等待完成**：main 函数退出时 goroutine 还在运行
3. **没有取消机制**：goroutine 无法被主动停止
4. **goroutine 泄漏**：channel 未关闭或 context 未取消导致 goroutine 永久阻塞

### 最佳实践

1. **始终使用 context 进行取消操作**：可以优雅地取消多个相关的 goroutine
2. **使用 WaitGroup 等待完成**：确保所有 goroutine 在退出前完成工作
3. **使用 worker pool 限制并发**：防止创建过多 goroutine 耗尽资源
4. **使用 channel 关闭作为信号**：通知多个 goroutine 同时退出
5. **在 defer 中调用 cancel()**：确保 context 被正确清理

> **关键原则**：Goroutine 必须有明确的生命周期管理策略，要么等待它完成，要么主动取消它，绝不能放任不管。
