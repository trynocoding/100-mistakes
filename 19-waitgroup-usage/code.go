package main

import (
	"fmt"
	"sync"
	"time"
)

// 模拟工作任务
func doWork(id int) {
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("goroutine %d 完成\n", id)
}

// =============================================================================
// 错误示例：在 goroutine 内部调用 Add(1)
//
// 问题：wg.Wait() 可能早于 goroutine 调用 wg.Add(1) 执行，
//       导致计数器仍为 0，Wait() 立即返回，goroutine 实际未执行完毕。
// =============================================================================
func incorrectUsage() {
	fmt.Println("=== 错误示例：在 goroutine 内调用 Add ===")

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		// 错误：在 goroutine 内部调用 Add(1)
		go func(id int) {
			wg.Add(1) // 问题：可能晚于 Wait() 执行
			doWork(id)
			wg.Done()
		}(i)
	}

	// 由于 Add 可能在 Wait() 之后执行，这里可能立即返回
	wg.Wait()
	fmt.Println("错误示例执行完毕（可能未等待所有 goroutine）\n")
}

// =============================================================================
// 正确示例：在启动 goroutine 前调用 Add(1)
//
// 正确做法：先调用 Add(1)，再启动 goroutine，确保计数器增加在 Wait() 之前。
// =============================================================================
func correctUsage() {
	fmt.Println("=== 正确示例：在 goroutine 外调用 Add ===")

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)           // 正确：在启动 goroutine 前调用
		go func(id int) {
			defer wg.Done() // 使用 defer 确保调用
			doWork(id)
		}(i)
	}

	wg.Wait()
	fmt.Println("正确示例执行完毕（已等待所有 goroutine）\n")
}

// =============================================================================
// 进阶：正确处理循环中的 goroutine 闭包变量
//
// 错误示例会展示闭包捕获问题——所有 goroutine 可能引用同一个变量。
// =============================================================================
func closureVariableDemo() {
	fmt.Println("=== 进阶：闭包变量捕获问题 ===")

	// 错误示例：闭包捕获的是变量的引用，而非值
	fmt.Println("错误示例：")
	var wg1 sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			// 由于闭包捕获的是 i 的引用，可能输出相同的值
			fmt.Printf("  错误方式输出: i = %d\n", i)
		}()
	}
	wg1.Wait()

	// 正确示例：通过参数传递当前值
	fmt.Println("正确示例：")
	var wg2 sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			// id 是当前迭代时传入的副本
			fmt.Printf("  正确方式输出: id = %d\n", id)
		}(i)
	}
	wg2.Wait()
	fmt.Println()
}

// =============================================================================
// 主函数：运行所有示例
// =============================================================================
func main() {
	fmt.Println("WaitGroup 正确用法演示\n")
	fmt.Println("Go 版本要求: Go 1.21+\n")

	incorrectUsage() // 先运行，观察输出

	time.Sleep(500 * time.Millisecond) // 等待 goroutine 打印完成

	correctUsage()

	time.Sleep(500 * time.Millisecond)

	closureVariableDemo()

	fmt.Println("所有演示完成。")
}
