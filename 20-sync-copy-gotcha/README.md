# 不要拷贝 sync 类型

## 问题描述

`sync.Mutex`、`sync.RWMutex`、`sync.WaitGroup`、`sync.Once`、`sync.Cond`、`sync.Map` 等 sync 包中的类型设计为值类型，但**不应被拷贝**。

拷贝这些类型会导致锁失效、计数混乱等严重问题，因为它们内部维护的状态是指向同一块内存的指针语义。

## 值拷贝导致的锁失效

以下示例展示了使用值 receiver 拷贝 `sync.Mutex` 导致的问题：

```go
package main

import (
	"fmt"
	"sync"
)

// Counter 使用值 receiver（错误示范）
type Counter struct {
	mu  sync.Mutex
cnt int
}

// Increment 值 receiver 拷贝了 mutex
func (c Counter) Increment() {
	c.mu.Lock()
	cnt++
	c.mu.Unlock()
}

// Value 获取计数（值 receiver）
func (c Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cnt
}

func main() {
	var counter Counter

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()
	fmt.Printf("期望值: 100, 实际值: %d\n", counter.Value())
}
```

运行结果：
```
期望值: 100, 实际值: 0
```

**原因分析**：`Increment` 和 `Value` 方法接收的是 `Counter` 的副本，它们各自的 `c.mu` 是不同的 mutex 对象。锁住 `c.mu` 只会影响当前方法内的副本，不会阻止其他副本的访问。

## 正确做法：使用指针 Receiver

```go
package main

import (
	"fmt"
	"sync"
)

// Counter 使用指针 receiver（正确做法）
type Counter struct {
	mu  sync.Mutex
	cnt int
}

// Increment 指针 receiver
func (c *Counter) Increment() {
	c.mu.Lock()
	cnt++
	c.mu.Unlock()
}

// Value 指针 receiver
func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cnt
}

func main() {
	counter := &Counter{}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()
	fmt.Printf("期望值: 100, 实际值: %d\n", counter.Value())
}
```

运行结果：
```
期望值: 100, 实际值: 100
```

**关键点**：
- 使用指针 receiver `(c *Counter)` 确保所有方法操作的是同一个 mutex 实例
- 创建对象时使用 `&Counter{}` 或 `new(Counter)` 获取指针

## 常见错误模式

### 错误 1：函数参数拷贝 mutex

```go
func (c *Counter) IncrementSafe(m sync.Mutex) { // 错误：参数拷贝
	m.Lock()
	c.cnt++
	m.Unlock()
}
```

### 错误 2：返回时拷贝 mutex

```go
func (c *Counter) GetMutex() sync.Mutex { // 错误：返回值拷贝
	return c.mu
}
```

### 错误 3：拷贝包含 sync 类型的结构体

```go
type Stats struct {
	mu    sync.Mutex
	Count int
}

func copyStats(s Stats) { // 错误：拷贝了 mu
	// s.mu 与原始的 mu 是不同的锁
}
```

## go vet 检测

`go vet` 可以自动检测这类问题：

```bash
$ go vet ./...
```

示例输出：
```
./main.go:14: Increment method has pointer receiver Counter but copies sync.Mutex value
./main.go:21: Value method has pointer receiver Counter but copies sync.Mutex value
```

启用严格检查：
```bash
go vet -copylocks ./...
```

## 总结

| 做法 | 结果 |
|------|------|
| 值 receiver `(c Counter)` | 每次调用拷贝 mutex，锁失效 |
| 指针 receiver `(c *Counter)` | 正确，共享同一个 mutex |

**最佳实践**：
1. 始终对包含 sync 类型的结构体使用指针 receiver
2. 永远不要将 sync 类型作为函数参数传递（拷贝问题）
3. 使用 `go vet` 持续检测此类问题
4. 优先将 sync 类型作为结构体的第一个字段，并使用指针管理其生命周期
