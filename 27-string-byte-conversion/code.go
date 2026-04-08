package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"
)

// 演示字符串与字节切片的转换方式

func main() {
	fmt.Println("=== 字符串与字节切片转换演示 ===")
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Println()

	// 基础转换演示
	demonstrateBasicConversions()

	// 内存分配演示
	demonstrateMemoryAllocation()

	// unsafe 转换演示
	demonstrateUnsafeConversions()

	// 旧版 unsafe 方式的问题
	demonstrateOldUnsafeProblems()

	// 安全使用 unsafe 的规则
	demonstrateSafeUnsafeUsage()

	// 性能测试
	runPerformanceTests()
}

// =============================================================================
// 基础转换演示
// =============================================================================

func demonstrateBasicConversions() {
	fmt.Println("【基础转换】传统方式")
	fmt.Println("--------------")

	// 字符串转字节切片
	s := "hello, world"
	b := []byte(s)
	fmt.Printf("字符串: %q\n", s)
	fmt.Printf("[]byte(s): %v\n", b)

	// 字节切片转字符串
	s2 := string(b)
	fmt.Printf("string(b): %q\n", s2)
	fmt.Println()

	// 验证数据独立
	b[0] = 'H'
	fmt.Printf("修改 b[0] 后，s 保持不变: %q\n", s)
	fmt.Printf("但 s2 也保持不变: %q\n", s2)
	fmt.Println("结论：传统转换会复制数据，各份数据独立")
	fmt.Println()
}

// =============================================================================
// 内存分配演示
// =============================================================================

func demonstrateMemoryAllocation() {
	fmt.Println("【内存分配】传统方式 vs unsafe 方式")
	fmt.Println("----------------------------------")

	// 测试字符串转字节切片
	s := strings.Repeat("A", 1000)

	// 传统方式
	alloc := testing.AllocsPerRun(1000, func() {
		_ = []byte(s)
	})
	fmt.Printf("[]byte(s) 平均分配: %.2f 次/操作\n", alloc)

	// unsafe 方式
	alloc = testing.AllocsPerRun(1000, func() {
		_ = unsafeSliceFromString(s)
	})
	fmt.Printf("unsafe.Slice(StringData(s), len(s)) 平均分配: %.2f 次/操作\n", alloc)
	fmt.Println()

	// 测试字节切片转字符串
	origB := bytes.Repeat([]byte{'X'}, 1000)

	alloc = testing.AllocsPerRun(1000, func() {
		_ = string(origB)
	})
	fmt.Printf("string(b) 平均分配: %.2f 次/操作\n", alloc)

	alloc = testing.AllocsPerRun(1000, func() {
		_ = unsafeStringFromSlice(origB)
	})
	fmt.Printf("unsafe.String(&b[0], len(b)) 平均分配: %.2f 次/操作\n", alloc)
	fmt.Println()
}

// =============================================================================
// unsafe 转换演示（Go 1.15+ 现代 API）
// =============================================================================

func demonstrateUnsafeConversions() {
	fmt.Println("【现代 unsafe API】Go 1.15+")
	fmt.Println("------------------------")

	s := "hello, unsafe"

	// 使用现代 unsafe API 转换
	b := unsafeSliceFromString(s)
	fmt.Printf("字符串: %q\n", s)
	fmt.Printf("转换为字节切片: %v\n", b)
	fmt.Printf("字节切片长度: %d\n", len(b))

	// 反向转换
	s2 := unsafeStringFromSlice(b)
	fmt.Printf("转回字符串: %q\n", s2)
	fmt.Printf("两者相等: %v\n", s == s2)
	fmt.Println()

	// 空字符串和空切片处理
	emptyB := unsafeSliceFromString("")
	fmt.Printf("空字符串转换: %v (len=%d)\n", emptyB, len(emptyB))

	emptyS := unsafeStringFromSlice([]byte{})
	fmt.Printf("空切片转换: %q (len=%d)\n", emptyS, len(emptyS))
	fmt.Println()
}

// =============================================================================
// 旧版 unsafe 方式的问题
// =============================================================================

func demonstrateOldUnsafeProblems() {
	fmt.Println("【旧版 unsafe 方式的问题】Go 1.12-1.14")
	fmt.Println("--------------------------------------")

	fmt.Println("问题一：依赖内部结构，无法保证跨版本兼容")
	fmt.Println("旧实现直接假设 string 和 slice 的内存布局：")
	fmt.Println("  string: { ptr *byte, len int }")
	fmt.Println("  slice:  { ptr *byte, len int, cap int }")
	fmt.Println()

	fmt.Println("问题二：修改转换后的切片会引发未定义行为")
	fmt.Println("字符串设计为不可变，但字节切片是可变的：")

	s := "hello"
	fmt.Printf("原始字符串: %q\n", s)

	// 危险操作：修改转换后的字节切片
	b := oldStringToBytes(s)
	fmt.Printf("转换后字节切片: %v\n", b)

	// 注意：这里不实际执行修改，因为结果未定义
	fmt.Println("修改 b[0] 会导致未定义行为（可能改变原始字符串）")
	fmt.Println("强烈不推荐在生产代码中这样做")
	fmt.Println()

	fmt.Println("问题三：切片操作可能导致内存泄漏")
	fmt.Println("转换后的切片引用原始内存，即使只使用一小部分，")
	fmt.Println("原始数据也无法被 GC 回收")
	fmt.Println()
}

// =============================================================================
// 安全使用 unsafe 的规则
// =============================================================================

func demonstrateSafeUnsafeUsage() {
	fmt.Println("【安全使用 unsafe 的规则】")
	fmt.Println("-------------------------")

	fmt.Println("规则一：字节数据必须来自可寻址的数组或切片")

	// 正确：使用堆上的切片
	arr := []byte{'h', 'e', 'l', 'l', 'o'}
	s := unsafe.String(&arr[0], len(arr))
	fmt.Printf("正确方式: %q\n", s)

	fmt.Println()
	fmt.Println("规则二：不要修改转换后字符串对应的底层字节")
	fmt.Println("（字符串在语义上是不可变的）")

	b := unsafeSliceFromString("hello")
	fmt.Printf("原始字节: %v\n", b)
	fmt.Println("尝试修改 b[0] 会导致数据竞争和未定义行为")

	fmt.Println()
	fmt.Println("规则三：确保数据生命周期覆盖使用范围")

	// 这个函数有问题：返回的字节切片引用了即将被回收的栈内存
	fmt.Println("错误示例（已注释，因为运行会出问题）:")
	fmt.Println("  func bad() []byte {")
	fmt.Println("      s := \"hello\"")
	fmt.Println("      return unsafeSliceFromString(s)")
	fmt.Println("  }")
	fmt.Println("  // s 在函数返回后生命周期结束，但返回值仍在使用它")
	fmt.Println()

	// 正确示例：数据来自调用者的参数或全局变量
	data := []byte("hello")
	result := safeConvert(data)
	fmt.Printf("正确示例: 数据来自参数: %q\n", result)
	fmt.Println()
}

// =============================================================================
// 辅助函数
// =============================================================================

// 现代 unsafe 方式：字符串转字节切片（Go 1.15+）
func unsafeSliceFromString(s string) []byte {
	if s == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// 现代 unsafe 方式：字节切片转字符串（Go 1.15+）
func unsafeStringFromSlice(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// 旧版 unsafe 方式（仅用于演示问题，不推荐使用）
func oldStringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// 旧版 unsafe 方式（仅用于演示问题，不推荐使用）
func oldBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// 安全示例：数据来自调用者参数
func safeConvert(b []byte) string {
	// 确保数据不会在函数返回后被回收
	// 这里使用副本确保安全
	clone := bytes.Clone(b)
	return unsafeStringFromSlice(clone)
}

// =============================================================================
// 性能测试
// =============================================================================

func runPerformanceTests() {
	fmt.Println("【性能测试】")
	fmt.Println("------------")

	// 字符串转字节切片性能对比
	fmt.Println("字符串转字节切片（1000次迭代）：")

	s := strings.Repeat("A", 1000)
	var b []byte

	// 传统方式
	start := time.Now()
	for i := 0; i < 1000; i++ {
		b = []byte(s)
	}
	elapsed := time.Since(start)
	fmt.Printf("  []byte(s):          %v\n", elapsed)

	// unsafe 方式
	start = time.Now()
	for i := 0; i < 1000; i++ {
		b = unsafeSliceFromString(s)
	}
	elapsed = time.Since(start)
	fmt.Printf("  unsafe.Slice():     %v\n", elapsed)

	// bytes.Clone 方式（创建独立副本）
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_ = bytes.Clone(b)
	}
	elapsed = time.Since(start)
	fmt.Printf("  bytes.Clone():      %v\n", elapsed)

	fmt.Println()

	// 字节切片转字符串性能对比
	fmt.Println("字节切片转字符串（1000次迭代）：")

	origB := bytes.Repeat([]byte{'X'}, 1000)
	var s2 string

	// 传统方式
	start = time.Now()
	for i := 0; i < 1000; i++ {
		s2 = string(origB)
	}
	elapsed = time.Since(start)
	fmt.Printf("  string(b):          %v\n", elapsed)

	// unsafe 方式
	start = time.Now()
	for i := 0; i < 1000; i++ {
		s2 = unsafeStringFromSlice(origB)
	}
	elapsed = time.Since(start)
	fmt.Printf("  unsafe.String():    %v\n", elapsed)

	// strings.Clone 方式
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_ = strings.Clone(s2)
	}
	elapsed = time.Since(start)
	fmt.Printf("  strings.Clone():    %v\n", elapsed)

	fmt.Println()
	fmt.Println("注意：实际性能取决于数据大小和 CPU 架构")
	fmt.Println("对于一般应用，传统方式的可读性和安全性更重要")
	fmt.Println()

	// 验证转换正确性
	fmt.Println("【正确性验证】")
	fmt.Println("--------------")

	testStrings := []string{"", "a", "hello", "hello, world", "你好, 世界", strings.Repeat("A", 1000)}

	for _, ts := range testStrings {
		// 往返转换验证
		b := []byte(ts)
		s := string(b)
		match := ts == s
		fmt.Printf("  %q -> []byte -> string: %v\n", truncateString(ts, 20), match)

		// unsafe 往返转换验证
		bUnsafe := unsafeSliceFromString(ts)
		sUnsafe := unsafeStringFromSlice(bUnsafe)
		matchUnsafe := ts == sUnsafe
		fmt.Printf("  %q -> unsafe -> string: %v\n", truncateString(ts, 20), matchUnsafe)
	}
	fmt.Println()
}

// truncateString 截断长字符串用于显示
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
