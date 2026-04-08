# WaitGroup 正确用法

## 概述

`sync.WaitGroup` 是 Go 中用于等待一组 goroutine 完成执行的常用原语。然而，如果使用不当，很容易引入竞态条件或导致死锁。

## 常见错误：在 goroutine 内部调用 Add(1)

### 问题分析

在 goroutine 内部调用 `wg.Add(1)` 是错误的，因为：

1. **时序不确定性**：goroutine 的启动和执行是异步的，主 goroutine 可能在新启动的 goroutine 真正执行 `Add` 之前就调用了 `wg.Wait()`。

2. **竞态条件**：如果 `Wait()` 在 `Add()` 之前被调用，计数器仍然是 0，`Wait()` 会立即返回，而实际的 goroutine 还未开始执行。

### 错误示例

```go
var wg sync.WaitGroup

func main() {
    for i := 0; i < 5; i++ {
        // 错误：在 goroutine 内部调用 Add
        go func() {
            wg.Add(1) // 问题：可能晚于 Wait() 执行
            doWork(i)
            wg.Done()
        }()
    }

    wg.Wait() // 可能立即返回，因为 Add 还未被调用
}
```

在上述代码中，`wg.Wait()` 可能在所有 goroutine 调用 `wg.Add(1)` 之前就执行完毕，导致 `doWork` 未被实际执行就被"等待"结束了。

## 正确做法：在启动 goroutine 之前调用 Add

### 核心原则

**始终在启动 goroutine 之前调用 `wg.Add(1)`**，然后在 goroutine 内部调用 `wg.Done()`。

### 正确示例

```go
var wg sync.WaitGroup

func main() {
    for i := 0; i < 5; i++ {
        wg.Add(1) // 正确：在启动 goroutine 前调用
        go func(id int) {
            defer wg.Done() // 使用 defer 确保即使发生 panic 也会调用
            doWork(id)
        }(i)
    }

    wg.Wait() // 等待所有 goroutine 完成
}
```

### 改进：错误示例的修正

```go
var wg sync.WaitGroup

func main() {
    for i := 0; i < 5; i++ {
        wg.Add(1) // 在 goroutine 启动前调用
        go func() {
            defer wg.Done()
            doWork(i)
        }()
    }

    wg.Wait()
}
```

## 进阶：正确处理循环中的 goroutine

### 闭包变量捕获问题

在使用闭包启动 goroutine 时，需要注意循环变量的捕获方式：

```go
// 错误：所有 goroutine 可能使用相同的变量值
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println(i) // i 引用的是同一个变量
    }()
}

// 正确：通过参数传递当前值
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        fmt.Println(id) // id 是当前迭代的副本
    }(i)
}
```

## 总结

| 错误做法 | 正确做法 |
|---------|---------|
| 在 goroutine 内调用 `Add` | 在启动 goroutine 前调用 `Add` |
| `Wait()` 可能立即返回 | `Wait()` 正确等待所有 goroutine |
| 不使用 `defer wg.Done()` | 使用 `defer wg.Done()` 确保调用 |

## 最佳实践

1. **先 Add，后 Go**：始终按照 `Add(1)` -> `go func()` -> `Done()` 的顺序编写代码。

2. **使用 defer**：在 goroutine 内部使用 `defer wg.Done()`，避免因提前返回或 panic 而遗漏。

3. **避免在循环中捕获闭包变量**：通过参数传递循环变量的当前值。

4. **注意 Add 和 Done 的数量匹配**：确保 `Add` 的调用次数等于 `Done` 的调用次数。

---

完整代码示例请参见 [code.go](./code.go)。
