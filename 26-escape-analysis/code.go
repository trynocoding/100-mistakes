package main

import (
	"fmt"
	"runtime"
)

// =============================================================================
// 逃逸分析（Escape Analysis）示例代码
// Go 1.21+ 可运行
//
// 运行方式：
//   go run -gcflags '-m -l' code.go
//
// 预期输出会显示哪些变量逃逸到堆上。
// =============================================================================

// -----------------------------------------------------------------------------
// 场景 1：返回指针 - 变量逃逸到堆
// -----------------------------------------------------------------------------

// returnPointer 返回一个指向局部变量的指针
// 编译时会输出：returnPointer new(int) escapes to heap
func returnPointer() *int {
	v := 10 // v 分配在栈上，但返回指针时 v 逃逸到堆
	return &v
}

// returnPointerFixed 通过参数接收指针，由调用者决定内存分配
// 不会产生逃逸
func returnPointerFixed(result *int) {
	*result = 10
}

// -----------------------------------------------------------------------------
// 场景 2：接口类型（any）返回 - 变量逃逸
// -----------------------------------------------------------------------------

// interfaceReturn 返回 any 类型
// 编译时会输出：interfaceReturn new(int) escapes to heap
func interfaceReturn() any {
	v := 10
	return v // v 逃逸到堆，因为返回的是接口类型
}

// concreteReturn 返回具体类型
// 不会逃逸（如果编译器能确定大小）
func concreteReturn() int {
	v := 10
	return v
}

// -----------------------------------------------------------------------------
// 场景 3：闭包捕获 - 捕获的变量逃逸
// -----------------------------------------------------------------------------

// closureCapture 返回一个闭包，闭包捕获了局部变量
// 编译时会输出：closureCapture.func1 v escapes to heap
func closureCapture() func() int {
	v := 10 // v 逃逸到堆，被闭包引用
	return func() int {
		return v
	}
}

// closureNoCapture 闭包不捕获外部变量
// v 不会逃逸
func closureNoCapture() func() int {
	fn := func() int { // 闭包不引用外部变量
		return 10
	}
	return fn
}

// -----------------------------------------------------------------------------
// 场景 4：切片长度/容量为变量
// -----------------------------------------------------------------------------

// sliceWithVariableLen 切片长度是变量
// 在 Go 1.21+ 中，如果 n 较小，可能会在栈上分配
// 编译时会输出关于逃逸的信息
func sliceWithVariableLen(n int) []int {
	// 当 n 是编译期常量时，编译器可能优化为栈分配
	// 当 n 是变量时，Go 1.21+ 会尽量栈分配
	return make([]int, n)
}

// sliceWithLargeCapacity 大容量切片可能会逃逸
func sliceWithLargeCapacity(n int) []int {
	return make([]int, 0, n) // 大容量切片可能会分配在堆上
}

// -----------------------------------------------------------------------------
// 场景 5：指针切片
// -----------------------------------------------------------------------------

// pointerSlice 返回指针切片
// 每个元素的指针指向堆上分配的值
func pointerSlice(n int) []*int {
	result := make([]*int, n)
	for i := 0; i < n; i++ {
		v := i // v 是新变量，每次循环都创建
		result[i] = &v // v 逃逸到堆
	}
	return result
}

// pointerSliceFixed 修复版本 - 直接分配在切片中
func pointerSliceFixed(n int) []int {
	result := make([]int, n)
	for i := 0; i < n; i++ {
		result[i] = i // 直接存储值，不产生逃逸
	}
	return result
}

// -----------------------------------------------------------------------------
// 场景 6：字符串拼接
// -----------------------------------------------------------------------------

// stringConcat 字符串拼接可能导致逃逸
// 在 Go 1.21+ 中，小规模拼接通常会被优化
func stringConcat(s1, s2, s3 string) string {
	return s1 + s2 + s3
}

// stringConcatWithInterface 使用接口拼接
// 编译时会输出：escape to heap
func stringConcatWithInterface(s1, s2, s3 any) string {
	return fmt.Sprintf("%v%v%v", s1, s2, s3)
}

// -----------------------------------------------------------------------------
// 场景 7：defer 和闭包
// -----------------------------------------------------------------------------

// deferWithClosure defer 使用闭包
// 编译时会输出：deferReturn func1 v escapes to heap
func deferWithClosure() int {
	v := 10
	defer func() {
		// v 逃逸到堆，因为 defer 语句引用了 v
		fmt.Println("defer:", v)
	}()
	return v
}

// deferWithoutClosure defer 不使用闭包
// v 不会逃逸
func deferWithoutClosure() int {
	v := 10
	defer fmt.Println("defer:", v) // v 作为参数传递，不逃逸
	return v
}

// -----------------------------------------------------------------------------
// 场景 8：map 和 channel
// -----------------------------------------------------------------------------

// mapAccess map 操作可能导致逃逸
func mapAccess() {
	m := make(map[string]int) // m 逃逸到堆
	m["key"] = 10
	_ = m
}

// channelSend 通过 channel 发送指针
// 指针指向的变量可能逃逸
func channelSend(ch chan<- *int) {
	v := 10
	ch <- &v // v 逃逸到堆，因为 channel 接收者可能不在同一函数
}

// -----------------------------------------------------------------------------
// 辅助函数：打印内存统计
// -----------------------------------------------------------------------------

func printMemStats(msg string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%s - Alloc: %d KB, TotalAlloc: %d KB, NumGC: %d\n",
		msg, m.Alloc/1024, m.TotalAlloc/1024, m.NumGC)
}

// -----------------------------------------------------------------------------
// 主函数：运行所有示例
// -----------------------------------------------------------------------------

func main() {
	fmt.Println("=== 逃逸分析示例 ===\n")

	// 场景 1：返回指针
	fmt.Println("--- 场景 1: 返回指针 ---")
	p := returnPointer()
	fmt.Printf("返回的指针值: %d\n", *p)
	fmt.Println()

	// 场景 2：接口返回
	fmt.Println("--- 场景 2: 接口类型返回 ---")
	i := interfaceReturn()
	fmt.Printf("返回的接口值: %v\n", i)
	j := concreteReturn()
	fmt.Printf("返回的具体类型值: %d\n", j)
	fmt.Println()

	// 场景 3：闭包捕获
	fmt.Println("--- 场景 3: 闭包捕获 ---")
	fn := closureCapture()
	fmt.Printf("闭包返回值: %d\n", fn())
	fmt.Println()

	// 场景 4：切片长度
	fmt.Println("--- 场景 4: 切片长度/容量为变量 ---")
	smallSlice := sliceWithVariableLen(5)
	fmt.Printf("小切片: %v\n", smallSlice)
	largeSlice := sliceWithLargeCapacity(100)
	fmt.Printf("大容量切片: %v\n", largeSlice)
	fmt.Println()

	// 场景 5：指针切片
	fmt.Println("--- 场景 5: 指针切片 ---")
	ptrs := pointerSlice(3)
	for i, p := range ptrs {
		fmt.Printf("ptrs[%d] = %d\n", i, *p)
	}
	fixed := pointerSliceFixed(3)
	fmt.Printf("修复后: %v\n", fixed)
	fmt.Println()

	// 场景 6：字符串拼接
	fmt.Println("--- 场景 6: 字符串拼接 ---")
	s := stringConcat("Hello", " ", "World")
	fmt.Printf("拼接结果: %s\n", s)
	s2 := stringConcatWithInterface("A", "B", "C")
	fmt.Printf("接口拼接: %s\n", s2)
	fmt.Println()

	// 场景 7：defer 和闭包
	fmt.Println("--- 场景 7: defer 和闭包 ---")
	deferWithClosure()
	deferWithoutClosure()
	fmt.Println()

	// 场景 8：map 和 channel
	fmt.Println("--- 场景 8: map 和 channel ---")
	mapAccess()
	ch := make(chan *int, 1)
	channelSend(ch)
	fmt.Printf("从 channel 接收: %d\n", <-ch)
	fmt.Println()

	// 打印内存统计
	printMemStats("最终内存统计")

	fmt.Println("\n=== 示例运行完成 ===")
	fmt.Println("\n提示：使用以下命令查看逃逸分析输出：")
	fmt.Println("  go run -gcflags '-m -l' code.go")
}