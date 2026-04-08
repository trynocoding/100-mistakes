package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// =============================================================================
// time.After 内存泄漏示例代码
// Go 1.21+ 可运行
// =============================================================================

// -----------------------------------------------------------------------------
// 错误示例：循环内使用 time.After 导致内存泄漏
// -----------------------------------------------------------------------------

func printMemDiff(allocBefore, allocAfter uint64) {
	if allocAfter >= allocBefore {
		fmt.Printf("内存增长: %d KB\n", (allocAfter-allocBefore)/1024)
	} else {
		fmt.Printf("内存减少: %d KB\n", (allocBefore-allocAfter)/1024)
	}
}

// leakExample 演示 time.After 在循环中的内存泄漏问题
func leakExample(count int) {
	fmt.Printf("开始泄漏演示，处理 %d 个元素...\n", count)

	// 记录初始内存状态
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	allocBefore := m1.Alloc

	for i := 0; i < count; i++ {
		select {
		case <-time.After(10 * time.Millisecond): // 每次迭代创建新定时器
			// 模拟处理
			_ = i * 2
		}
	}

	// 强制 GC
	runtime.GC()
	runtime.ReadMemStats(&m1)
	allocAfter := m1.Alloc

	fmt.Printf("泄漏前内存: %d KB\n", allocBefore/1024)
	fmt.Printf("泄漏后内存: %d KB\n", allocAfter/1024)
	printMemDiff(allocBefore, allocAfter)
	fmt.Println("注意：如果 count 很大，内存会持续增长而无法回收")
}

// -----------------------------------------------------------------------------
// 正确示例 1：使用 time.NewTimer() + Reset() 复用定时器
// -----------------------------------------------------------------------------

// fixedWithTimer 演示使用 NewTimer + Reset 避免泄漏
func fixedWithTimer(count int) {
	fmt.Printf("使用 NewTimer + Reset 处理 %d 个元素...\n", count)

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	allocBefore := m1.Alloc

	// 创建一个定时器，复用同一个对象
	timer := time.NewTimer(10 * time.Millisecond)
	defer timer.Stop()

	for i := 0; i < count; i++ {
		// 重置定时器之前，先停止（如果已触发需要先消费）
		if !timer.Stop() {
			select {
			case <-timer.C: // 消耗已触发的信号
			default:
			}
		}
		timer.Reset(10 * time.Millisecond)

		select {
		case <-timer.C:
			_ = i * 2
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m1)
	allocAfter := m1.Alloc

	fmt.Printf("修复后泄漏前内存: %d KB\n", allocBefore/1024)
	fmt.Printf("修复后泄漏后内存: %d KB\n", allocAfter/1024)
	printMemDiff(allocBefore, allocAfter)
}

// -----------------------------------------------------------------------------
// 正确示例 2：使用 time.Ticker 处理周期性任务
// -----------------------------------------------------------------------------

// fixedWithTicker 演示使用 Ticker 处理周期性任务
func fixedWithTicker(count int) {
	fmt.Printf("使用 Ticker 处理 %d 个元素...\n", count)

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)
	allocBefore := m1.Alloc

	// 创建 Ticker，复用同一个对象
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	processed := 0
	for processed < count {
		select {
		case <-ticker.C:
			_ = processed * 2
			processed++
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m1)
	allocAfter := m1.Alloc

	fmt.Printf("Ticker 泄漏前内存: %d KB\n", allocBefore/1024)
	fmt.Printf("Ticker 泄漏后内存: %d KB\n", allocAfter/1024)
	printMemDiff(allocBefore, allocAfter)
}

// -----------------------------------------------------------------------------
// 完整示例：带 context 的心跳检测
// -----------------------------------------------------------------------------

// checkHealth 模拟健康检查
func checkHealth() error {
	time.Sleep(1 * time.Millisecond) // 模拟检查耗时
	return nil
}

// heartbeatWithAfter 错误：循环内使用 time.After
func heartbeatWithAfter(ctx context.Context, duration time.Duration) {
	fmt.Println("开始 heartbeatWithAfter（错误方式）...")

	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		select {
		case <-time.After(50 * time.Millisecond): // 每次循环创建新定时器
			if err := checkHealth(); err != nil {
				fmt.Printf("健康检查失败: %v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
	fmt.Println("heartbeatWithAfter 结束")
}

// heartbeatWithTicker 正确：使用 time.Ticker
func heartbeatWithTicker(ctx context.Context, duration time.Duration) {
	fmt.Println("开始 heartbeatWithTicker（正确方式）...")

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		select {
		case <-ticker.C:
			if err := checkHealth(); err != nil {
				fmt.Printf("健康检查失败: %v\n", err)
			}
		case <-ctx.Done():
			return
		}
	}
	fmt.Println("heartbeatWithTicker 结束")
}

// -----------------------------------------------------------------------------
// 演示 Reset 的正确用法
// -----------------------------------------------------------------------------

// resetExample 演示 Reset 的正确和错误用法
func resetExample() {
	fmt.Println("\n=== Reset 用法示例 ===")

	// 正确用法
	timer := time.NewTimer(100 * time.Millisecond)

	// 模拟在定时器触发前重置
	time.Sleep(50 * time.Millisecond)

	// 必须先 Stop，如果返回 false 说明已触发，需要消费通道
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	// 现在可以安全 Reset
	timer.Reset(100 * time.Millisecond)

	fmt.Println("定时器已重置")

	// 等待新的定时器触发
	<-timer.C
	fmt.Println("新的定时器触发")

	timer.Stop()
}

// -----------------------------------------------------------------------------
// 主函数
// -----------------------------------------------------------------------------

func main() {
	fmt.Println("=== time.After 内存泄漏演示 ===\n")

	// 1. 演示内存泄漏
	fmt.Println("--- 内存泄漏对比 ---\n")

	count := 100
	leakExample(count)
	fmt.Println()

	fixedWithTimer(count)
	fmt.Println()

	fixedWithTicker(count)
	fmt.Println()

	// 2. 演示心跳检测
	fmt.Println("--- 心跳检测对比 ---\n")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// 错误方式
	go heartbeatWithAfter(ctx, 200*time.Millisecond)

	// 等待上一个完成
	time.Sleep(250 * time.Millisecond)

	// 正确方式
	ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	heartbeatWithTicker(ctx2, 200*time.Millisecond)
	cancel2()

	// 3. Reset 用法
	resetExample()

	fmt.Println("\n=== 演示完成 ===")
	fmt.Println("\n关键点：")
	fmt.Println("1. time.After 在循环中每次都创建新定时器，导致内存泄漏")
	fmt.Println("2. time.NewTimer + Reset 可以复用同一个定时器")
	fmt.Println("3. time.Ticker 适合周期性任务，自动复用")
	fmt.Println("4. Reset 前需要先 Stop 并消费已触发的通道")
}