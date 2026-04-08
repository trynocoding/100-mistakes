package main

import (
	"fmt"
	"sync"
)

// ============================================================
// 错误示例：使用值 receiver 导致锁失效
// ============================================================

// BadCounter 使用值 receiver（错误示范）
type BadCounter struct {
	mu  sync.Mutex
	cnt int
}

// Increment 值 receiver 拷贝了 mutex，锁完全失效
func (c BadCounter) Increment() {
	c.mu.Lock()   // 这里锁住的是 c 的副本的 mutex
	c.cnt++       // 不是原对象的 mutex
	c.mu.Unlock() // 解锁的是副本的 mutex
}

// Value 值 receiver
func (c BadCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cnt
}

// ============================================================
// 正确示例：使用指针 receiver
// ============================================================

// GoodCounter 使用指针 receiver（正确做法）
type GoodCounter struct {
	mu  sync.Mutex
	cnt int
}

// Increment 指针 receiver
func (c *GoodCounter) Increment() {
	c.mu.Lock()
	c.cnt++
	c.mu.Unlock()
}

// Value 指针 receiver
func (c *GoodCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cnt
}

func main() {
	fmt.Println("=== sync 类型拷贝问题演示 ===")
	fmt.Println()

	// 测试错误示例
	badCounter := BadCounter{}
	var badWg sync.WaitGroup

	for i := 0; i < 100; i++ {
		badWg.Add(1)
		go func() {
			defer badWg.Done()
			badCounter.Increment()
		}()
	}

	badWg.Wait()
	fmt.Printf("错误示例（值 receiver）：\n")
	fmt.Printf("  期望值: 100, 实际值: %d\n", badCounter.Value())
	fmt.Printf("  问题：每次调用拷贝了 mutex，锁完全失效\n")
	fmt.Println()

	// 测试正确示例
	goodCounter := &GoodCounter{}
	var goodWg sync.WaitGroup

	for i := 0; i < 100; i++ {
		goodWg.Add(1)
		go func() {
			defer goodWg.Done()
			goodCounter.Increment()
		}()
	}

	goodWg.Wait()
	fmt.Printf("正确示例（指针 receiver）：\n")
	fmt.Printf("  期望值: 100, 实际值: %d\n", goodCounter.Value())
	fmt.Printf("  正确：所有方法共享同一个 mutex\n")
}
