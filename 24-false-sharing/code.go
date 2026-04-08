package main

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

func main() {
	RunDemo()
}
