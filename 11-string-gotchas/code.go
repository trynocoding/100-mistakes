package main

import (
	"fmt"
	"runtime"
	"strings"
	"unicode/utf8"
)

// 演示 String 处理的常见陷阱

func main() {
	fmt.Println("=== String 处理陷阱演示 ===")
	fmt.Println()

	// 陷阱一：字节索引 vs 字符索引
	fmt.Println("【陷阱一】字节索引 vs 字符索引")
	{
		s := "Hello世界"

		fmt.Printf("字符串: %q\n", s)
		fmt.Printf("len(s) = %d (字节数)\n", len(s))
		fmt.Printf("utf8.RuneCountInString(s) = %d (字符数)\n", utf8.RuneCountInString(s))
		fmt.Println()

		fmt.Println("按字节索引遍历（错误方式）:")
		for i := 0; i < len(s); i++ {
			fmt.Printf("  s[%d] = 0x%02x (%c)\n", i, s[i], s[i])
		}
		fmt.Println()

		fmt.Println("按 Unicode 字符遍历（正确方式）:")
		for i, r := range s {
			fmt.Printf("  s[%d] = %c (0x%04x)\n", i, r, r)
		}
		fmt.Println()

		// 错误访问
		fmt.Println("错误访问 s[6]:")
		r, _ := utf8.DecodeRuneInString(s[6:])
		fmt.Printf("  s[6] 字节值: 0x%02x, 解析为字符: %c\n", s[6], r)

		// 正确访问
		fmt.Println("\n正确访问第 7 个字符（索引从 0 开始）:")
		runes := []rune(s)
		fmt.Printf("  runes[6] = %c\n", runes[6])
	}
	fmt.Println()

	// 陷阱二：切片截断与内存泄漏
	fmt.Println("【陷阱二】切片截断与内存泄漏")
	{
		// 为了演示，我们使用较小的数据
		largeData := make([]byte, 64)
		for i := range largeData {
			largeData[i] = 'A'
		}
		large := string(largeData)

		fmt.Printf("原始字符串长度: %d 字节\n", len(large))

		// 切片：small 引用 large 的底层数组
		small := large[:8]
		fmt.Printf("切片后 small 长度: %d 字节\n", len(small))
		fmt.Printf("small: %q\n", small)

		// 如果 large 很大，即使 small 很小，large 仍然无法被 GC 回收
		// 因为 small 持有对底层数组的引用

		// 正确做法：使用 strings.Clone 创建独立副本
		cloned := strings.Clone(small)
		fmt.Printf("克隆后 cloned: %q\n", cloned)
		fmt.Println("strings.Clone 创建了独立副本，large 可以被垃圾回收")
	}
	fmt.Println()

	// 陷阱三：strings.Clone vs 手动复制
	fmt.Println("【陷阱三】strings.Clone vs 手动复制")
	{
		s := "Hello, 世界！这是一个测试字符串。"

		// 方式一：手动复制（传统方式）
		s1 := string([]byte(s))

		// 方式二：strings.Clone（Go 1.18+）
		s2 := strings.Clone(s)

		fmt.Printf("原始字符串: %q\n", s)
		fmt.Printf("string([]byte(s)): %q\n", s1)
		fmt.Printf("strings.Clone(s): %q\n", s2)
		fmt.Printf("两者相等: %v\n", s1 == s2)
		fmt.Println()
		fmt.Println("strings.Clone 的优势:")
		fmt.Println("  1. 减少一次内存分配")
		fmt.Println("  2. 避免临时 []byte 对象")
		fmt.Println("  3. 代码意图更清晰")
	}
	fmt.Println()

	// 陷阱四：字符串拼接
	fmt.Println("【陷阱四】字符串拼接")
	{
		strs := []string{"Go", "是", "一", "门", "好", "语言"}

		// 方式一：使用 + （低效）
		var result1 string
		for _, s := range strs {
			result1 += s
		}

		// 方式二：使用 strings.Builder （高效）
		var b strings.Builder
		for _, s := range strs {
			b.WriteString(s)
		}
		result2 := b.String()

		// 方式三：strings.Join （最简洁）
		result3 := strings.Join(strs, "")

		fmt.Printf("使用 + 拼接: %q\n", result1)
		fmt.Printf("使用 strings.Builder: %q\n", result2)
		fmt.Printf("使用 strings.Join: %q\n", result3)
		fmt.Printf("结果都相等: %v\n", result1 == result2 && result2 == result3)
	}
	fmt.Println()

	// 综合示例：常见错误与纠正
	fmt.Println("【综合示例】常见错误模式")
	{
		s := "你好世界"

		fmt.Println("错误模式 1: 使用字节索引访问字符")
		fmt.Printf("  s[0] = 0x%02x (可能只是'你'的一部分)\n", s[0])

		fmt.Println("正确模式 1: 使用 range 遍历")
		for i, r := range s {
			fmt.Printf("  索引 %d 处字符: %c\n", i, r)
		}

		fmt.Println()

		fmt.Println("错误模式 2: 切片大字符串")
		large := strings.Repeat("ABC", 100)
		small := large[:10]
		fmt.Printf("  small = %q 但引用了 large 的全部内存\n", small)

		fmt.Println("正确模式 2: 切片后克隆")
		smallCloned := strings.Clone(large[:10])
		fmt.Printf("  smallCloned = %q 是独立副本\n", smallCloned)
	}
	fmt.Println()

	// 演示 GC 前的内存状态
	fmt.Printf("当前 Go 版本运行 GC: runtime.Version() = %s\n", runtime.Version())
	fmt.Println("注意: strings.Clone 可确保切片操作不会意外保留大字符串在内存中")
}
