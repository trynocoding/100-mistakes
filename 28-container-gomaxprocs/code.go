//go:build ignore

package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"go.uber.org/automaxprocs/maxprocs"
)

func main() {
	// 演示问题：不使用 maxprocs 时，GOMAXPROCS 返回宿主机的核数
	fmt.Println("=== 问题演示：未使用 maxprocs ===")
	fmt.Printf("runtime.NumCPU(): %d\n", runtime.NumCPU())
	fmt.Printf("runtime.GOMAXPROCS(0): %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	// 使用 maxprocs.Set 自动设置正确的 GOMAXPROCS
	fmt.Println("=== 使用 maxprocs.Set 自动设置 ===")
	// Set 会自动读取 cgroup 配置并设置 GOMAXPROCS
	// 使用 Logger 选项可以看到设置日志
	undo, err := maxprocs.Set(maxprocs.Logger(log.Printf))
	if err != nil {
		// 在非容器环境或无法获取 cgroup 信息时，err 可能不为 nil
		fmt.Printf("maxprocs warning: %v\n", err)
	}
	defer undo() // 可以通过 undo() 恢复之前的设置

	fmt.Printf("自动设置后 runtime.GOMAXPROCS(0): %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	// 模拟工作负载，观察调度行为
	fmt.Println("=== 工作负载演示 ===")
	simulateWorkload()
}

func simulateWorkload() {
	const goroutineCount = 10
	done := make(chan struct{}, goroutineCount)

	start := time.Now()
	for i := 0; i < goroutineCount; i++ {
		go func(id int) {
			// 模拟 CPU 密集型工作
			_ = computeSomething()
			done <- struct{}{}
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < goroutineCount; i++ {
		<-done
	}

	fmt.Printf("完成 %d 个 goroutine，耗时: %v\n", goroutineCount, time.Since(start))
	fmt.Printf("当前 GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("当前 Goroutine 数量: %d\n", runtime.NumGoroutine())
}

func computeSomething() int {
	// 模拟一些计算工作
	const iterations = 1000000
	sum := 0
	for i := 0; i < iterations; i++ {
		sum += i % 10
	}
	return sum
}
