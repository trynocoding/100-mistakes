# Happens-Before 保证

## 概述

Happens-Before 是 Go 内存模型中定义的一种偏序关系，它定义了 goroutine 之间共享变量时的可见性保证。理解 Happens-Before 对于正确编写并发代码至关重要。

---

## 1. Goroutine 创建先于执行

**规则**：在 goroutine 开始执行之前，其创建语句一定已经执行完毕。

```go
var msg string

func main() {
    go func() {
        // 此处读取 msg 一定能看到 "hello"
        println(msg)
    }()
    msg = "hello"
    time.Sleep(time.Second)
}
```

**解释**：创建 goroutine 的代码 `go func(){}()` 会在新 goroutine 开始执行之前完成。因此，`msg = "hello"` 有可能在新 goroutine 之前执行，也有可能之后执行（调度不确定性）。

---

## 2. Goroutine 退出无保证

**规则**：主 goroutine 退出时，不会等待其他 goroutine 完成。

```go
func main() {
    go func() {
        time.Sleep(time.Second)
        println("goroutine done")
    }()
    println("main exit")
    // 程序直接退出，不会等待上面的 goroutine
}
```

**解释**：主函数返回后，程序会立即终止，不管是否有其他 goroutine 还在运行。

---

## 3. Channel Send 先于 Receive

**规则**：对同一个 channel 的发送操作一定happens-before接收操作完成。

```go
ch := make(chan int)

go func() {
    ch <- 1 // Send
}()

go func() {
    val := <-ch // Receive
    println(val) // 一定能收到 1
}()
```

**解释**：发送操作会在接收操作准备好之前完成（对于无缓冲 channel）。

---

## 4. Channel Close 先于 Receive

**规则**：关闭 channel 的操作一定happens-before接收操作收到零值。

```go
ch := make(chan int)

go func() {
    close(ch)
}()

go func() {
    val, ok := <-ch
    // val=0, ok=false 表示 channel 已关闭
    println(val, ok)
}()
```

**解释**：关闭操作完成之后，接收操作才会完成。

---

## 5. Unbuffered Channel: Receive 先于 Send

**规则**：对于无缓冲 channel，接收操作一定happens-before发送操作完成。

```go
ch := make(chan int)

go func() {
    ch <- 1 // Send 会等待直到有接收者
    println("send done")
}()

go func() {
    val := <-ch // Receive
    println("received:", val)
}()
```

**解释**：发送方会阻塞直到有接收方接收，这个同步机制使得无缓冲 channel 天然用于 goroutine 间的同步。

---

## Happens-Before 汇总表

| 场景 | 保证 |
|------|------|
| Goroutine 创建 | 创建 happens-before 开始执行 |
| Goroutine 退出 | 无保证，主 goroutine 退出不等待其他 goroutine |
| Channel Send | Send happens-before Receive 完成 |
| Channel Close | Close happens-before 接收返回零值 |
| Unbuffered Channel Receive | Receive happens-before Send 完成 |

---

## 最佳实践

1. **使用 channel 进行同步**：无缓冲 channel 提供了天然的同步机制
2. **使用 sync.WaitGroup 等待 goroutine**：如果需要等待多个 goroutine 完成
3. **避免依赖 goroutine 执行顺序**：调度是不确定的
4. **使用 buffered channel 时注意**：只有容量耗尽或关闭时才会阻塞

---

## 延伸阅读

- [Go 内存模型官方文档](https://go.dev/ref/mem)
- [Go并发模式](https://go.dev/blog/pipelines)
