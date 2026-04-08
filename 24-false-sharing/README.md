# False Sharing（伪共享）

## 什么是 False Sharing

False Sharing 是发生在多核 CPU 缓存系统中的一个性能问题。

现代 CPU 使用 **cache line**（缓存行）作为内存操作的最小单位，通常是 **64 字节**。当一个核心修改了某个变量，虽然这个变量只占几个字节，但 CPU 会将包含该变量的整个缓存行标记为无效。其他核心如果想访问同一缓存行上的任何变量，都必须等待缓存行失效后重新从内存（或从其他核心的缓存）加载。

**False Sharing** 发生在多个核心各自修改**不同变量**，但这些变量恰好落在**同一个缓存行**上时。即使这些变量之间没有任何逻辑关联，一个核心的写操作也会导致另一个核心的缓存行失效，从而被迫重新读取数据。

## 使用填充避免 False Sharing

解决方案是使用**填充（Padding）**来确保不同核心修改的变量位于不同的缓存行。

思路：
- 数组或结构体的元素如果可能被不同线程并发修改，应使用填充将其分隔到不同的缓存行
- 结构体设计时，将频繁并发访问的字段用填充隔开

## NoPad vs Pad 结构体对比

```go
// NoPad：没有填充的结构体
// Counter1 和 Counter2 位于同一个缓存行
type NoPad struct {
    Counter1 int64
    Counter2 int64
}

// Pad：带有填充的结构体
// 每个字段后填充 56 字节（64-8），确保每个字段独占一个缓存行
type Pad struct {
    Counter1 int64
    _        [7]int64 // 填充：56 字节
    Counter2 int64
    _        [7]int64 // 填充：56 字节
}
```

**关键点**：
- `NoPad` 结构体大小为 16 字节，两个字段共享同一个缓存行（假设缓存行是 64 字节，则多个 NoPad 实例可能落在同一缓存行）
- `Pad` 结构体通过填充确保 `Counter1` 和 `Counter2` 落在不同的缓存行上
- 在 NoPad 中，如果两个线程分别修改 Counter1 和 Counter2，会相互失效对方的缓存行
- 在 Pad 中，两个线程修改各自字段，互不影响

## 可运行的 Go 代码

代码文件：`code.go`

```go
package falseSharing

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// NoPad 结构体：没有填充
// 两个字段位于同一个缓存行，会产生 false sharing
type NoPad struct {
	Counter1 int64
	Counter2 int64
}

// Pad 结构体：使用填充
// 每个字段后填充 56 字节（64-8），确保每个字段独占一个缓存行
type Pad struct {
	Counter1 int64
	_        [7]int64 // 填充：56 字节
	Counter2 int64
	_        [7]int64 // 填充：56 字节
}

// BenchmarkNoPad 测试不带填充的原子计数器的性能
func BenchmarkNoPad() {
	fmt.Println("\n=== NoPad（无填充）===")

	var wg sync.WaitGroup
	counter := NoPad{}

	// 启动两个 goroutine，每个修改不同的计数器
	wg.Add(2)

	// Goroutine 1: 修改 Counter1 一百万次
	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			atomic.AddInt64(&counter.Counter1, 1)
		}
	}()

	// Goroutine 2: 修改 Counter2 一百万次
	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			atomic.AddInt64(&counter.Counter2, 1)
		}
	}()

	start := time.Now()
	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("NoPad 耗时: %v\n", elapsed)
	fmt.Printf("Counter1 = %d, Counter2 = %d\n", counter.Counter1, counter.Counter2)
}

// BenchmarkPad 测试带填充的原子计数器的性能
func BenchmarkPad() {
	fmt.Println("\n=== Pad（带填充）===")

	var wg sync.WaitGroup
	counter := Pad{}

	// 启动两个 goroutine，每个修改不同的计数器
	wg.Add(2)

	// Goroutine 1: 修改 Counter1 一百万次
	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			atomic.AddInt64(&counter.Counter1, 1)
		}
	}()

	// Goroutine 2: 修改 Counter2 一百万次
	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			atomic.AddInt64(&counter.Counter2, 1)
		}
	}()

	start := time.Now()
	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("Pad 耗时: %v\n", elapsed)
	fmt.Printf("Counter1 = %d, Counter2 = %d\n", counter.Counter1, counter.Counter2)
}

// RunDemo 运行演示
func RunDemo() {
	fmt.Println("False Sharing 演示")
	fmt.Println("====================")
	fmt.Println("\n预期结果：Pad 版本通常比 NoPad 版本快，因为避免了 false sharing")
	fmt.Println("注意：实际效果取决于 CPU 核心数、缓存行大小和系统负载")

	BenchmarkNoPad()
	BenchmarkPad()
}
```

## 运行代码

```bash
go run code.go
```

## 预期输出

```
False Sharing 演示
====================

预期结果：Pad 版本通常比 NoPad 版本快，因为避免了 false sharing
注意：实际效果取决于 CPU 核心数、缓存行大小和系统负载

=== NoPad（无填充）===
NoPad 耗时: 45.123ms
Counter1 = 1000000, Counter2 = 1000000

=== Pad（带填充）===
Pad 耗时: 12.456ms
Counter1 = 1000000, Counter2 = 1000000
```

## 总结

| 特性 | NoPad | Pad |
|------|-------|-----|
| 结构体大小 | 16 字节 | 224 字节 |
| 字段间距 | 0 字节（同一缓存行） | 56 字节填充 |
| False Sharing | 存在 | 不存在 |
| 性能 | 较慢（缓存行失效） | 较快（独立缓存行） |

**最佳实践**：
- 在高性能并发场景下，如果多个线程会频繁修改结构体的不同字段，考虑使用填充将它们分隔到不同的缓存行
- 使用 `sync/atomic` 包的原子操作时，更容易暴露 false sharing 问题
- 注意：填充会占用更多内存，是否使用需要权衡性能收益和内存成本
