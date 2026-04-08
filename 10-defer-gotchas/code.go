// code.go - Defer 注意事项演示
//
// 运行方式: go run code.go
//
// 本文件演示 defer 的三个关键注意点：
// 1. 函数永不返回时 defer 永不执行
// 2. defer 参数求值时机（声明时计算）
// 3. 闭包中捕获变量的不同行为

package main

import (
	"fmt"
	"runtime"
	"time"
)

// =============================================================================
// 注意点 1: 函数永不返回时，defer 永不执行
// =============================================================================

// exampleNeverReturns 演示：当函数永不返回时，defer 不会执行
// 注意：死循环导致函数永不返回，defer 不会执行
func exampleNeverReturns() {
	defer fmt.Println("[1-defer] 这行不会打印 - 函数进入死循环")

	fmt.Println("[1] 开始执行...进入死循环")

	// 死循环 - 函数永不返回，defer 不会执行
	for {
		// 模拟永不返回的场景（例如等待某个永远不会发生的事件）
		time.Sleep(time.Hour) // 使用 Sleep 避免 CPU 100%
	}
}

// exampleNormalReturn 演示：正常返回时 defer 执行
func exampleNormalReturn() {
	defer fmt.Println("[1-defer-normal] 清理资源 - 函数正常返回")

	fmt.Println("[1] 正常返回示例完成")
}

// exampleGoexitDemonstration 演示 runtime.Goexit() 的行为
// 注意：Goexit 会终止当前 goroutine，但在终止前会执行已注册的 defer
// Go 1.21+ 行为：Goexit 在终止 goroutine 之前会执行所有已注册的 defer
func exampleGoexitDemonstration() {
	defer fmt.Println("[1-goexit-defer] 这个 defer 会在 Goexit 前执行")

	fmt.Println("[1-goexit] 即将调用 Goexit()...")
	runtime.Goexit()
	// Goexit 之后的代码永远不会执行
	fmt.Println("[1-goexit] 这行永远不会打印")
}

// =============================================================================
// 注意点 2: defer 参数求值时机（声明时计算）
// =============================================================================

func deferredParameterEvaluation() {
	fmt.Println("\n[2] === defer 参数求值时机 ===")

	x := 1
	defer fmt.Println("[2-defer] x 的值（声明时求值）:", x)

	x = 2
	fmt.Println("[2] 修改后 x =", x)
	// 输出: 修改后 x = 2
	// defer 执行时输出: x 的值（声明时求值）: 1
}

func deferredClosureEvaluation() {
	fmt.Println("\n[2] === 使用闭包捕获变量引用 ===")

	x := 1
	defer func() {
		fmt.Println("[2-defer-closure] x 的值（闭包执行时求值）:", x)
	}()

	x = 2
	fmt.Println("[2] 修改后 x =", x)
	// 输出: 修改后 x = 2
	// defer 执行时输出: x 的值（闭包执行时求值）: 2
}

// =============================================================================
// 注意点 3: 闭包中捕获变量的不同行为
// =============================================================================

// 注意：在 Go 1.22+ 中，循环变量在每次迭代时会重新创建，
// 所以直接使用循环变量不会出现问题。以下示例使用显式共享变量来演示陷阱。

func closureVariableCaptureWrong() {
	fmt.Println("\n[3] === 闭包捕获错误示例（按引用捕获）===")

	// 错误示例：使用共享变量
	var sharedFuncs []func()
	sharedVar := 0

	for sharedVar < 3 {
		sharedFuncs = append(sharedFuncs, func() {
			fmt.Printf("[3-wrong] sharedVar = %d\n", sharedVar)
		})
		sharedVar++
	}

	fmt.Println("[3-wrong] 调用闭包:")
	for _, f := range sharedFuncs {
		f()
	}
	// 输出: sharedVar = 3, sharedVar = 3, sharedVar = 3
	// 所有闭包都引用同一个 sharedVar，循环结束后 sharedVar = 3
}

func closureVariableCaptureCorrect() {
	fmt.Println("\n[3] === 闭包捕获正确示例（按值传递）===")

	funcs := make([]func(), 3)
	for i := 0; i < 3; i++ {
		// 方法 1：通过函数参数按值传递
		funcs[i] = func(val int) func() {
			return func() {
				fmt.Printf("[3-correct] val = %d\n", val)
			}
		}(i)

		// 方法 2：创建局部变量
		// i := i
		// funcs[i] = func() { fmt.Printf("[3-correct] i = %d\n", i) }
	}

	fmt.Println("[3-correct] 调用闭包:")
	for _, f := range funcs {
		f()
	}
	// 输出: val = 0, val = 1, val = 2
	// 每个闭包捕获的是当时的值拷贝
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("========================================")
	fmt.Println("Defer 注意事项演示")
	fmt.Println("========================================")

	// 注意点 1: 函数永不返回时 defer 永不执行
	fmt.Println("\n=== 注意点 1: 函数永不返回时 defer 永不执行 ===")

	// 正常返回的情况 - defer 会执行
	exampleNormalReturn()

	// 死循环永不返回的情况 - defer 不会执行
	// 使用单独的 goroutine 以免整个程序卡住
	go func() {
		exampleNeverReturns() // 死循环，永远不会返回
		// 下面的代码永远不会执行
		fmt.Println("[1] 这行永远不会打印")
	}()

	// 给 goroutine 一点时间启动
	time.Sleep(10 * time.Millisecond)
	fmt.Println("[1] 由于 exampleNeverReturns 永不返回，其 defer 不会执行")

	// Goexit 演示：Goexit 会执行已注册的 defer，但终止后续所有代码
	fmt.Println("\n[1] 演示 Goexit 的行为...")
	go exampleGoexitDemonstration()
	time.Sleep(10 * time.Millisecond)
	fmt.Println("[1] Goexit 执行了 defer，然后终止了 goroutine")

	// 注意点 2: defer 参数求值时机
	deferredParameterEvaluation()
	deferredClosureEvaluation()

	// 注意点 3: 闭包中捕获变量的不同行为
	closureVariableCaptureWrong()
	closureVariableCaptureCorrect()

	fmt.Println("\n========================================")
	fmt.Println("演示完成")
	fmt.Println("========================================")
}
