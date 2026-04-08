package main

import (
	"fmt"
	"time"
)

// Fields 结构体：包含两个 int64 字段
// 在内存中，A 和 B 字段交错排列
type Fields struct {
	A int64
	B int64
}

// BenchmarkSequentialAccess 测试顺序访问结构体切片中同一字段的性能
// 访问模式：fields[0].A, fields[1].A, fields[2].A, ...
func BenchmarkSequentialAccess() int64 {
	const size = 100_000_000
	slice := make([]Fields, size)

	// 初始化
	for i := 0; i < size; i++ {
		slice[i].A = int64(i)
		slice[i].B = int64(i)
	}

	sum := int64(0)
	start := time.Now()

	// 顺序访问所有元素的 A 字段
	for i := 0; i < size; i++ {
		sum += slice[i].A
	}

	elapsed := time.Since(start)
	fmt.Printf("顺序访问（[]Fields，访问 A 字段）: %v, sum=%d\n", elapsed, sum)
	return int64(elapsed)
}

// BenchmarkRandomAccess 测试"伪随机"访问结构体切片中不同字段的性能
// 访问模式：fields[0].A, fields[1].B, fields[2].A, fields[3].B, ...
// 这模拟了交替访问结构体不同字段的场景
func BenchmarkRandomAccess() int64 {
	const size = 100_000_000
	slice := make([]Fields, size)

	// 初始化
	for i := 0; i < size; i++ {
		slice[i].A = int64(i)
		slice[i].B = int64(i)
	}

	sum := int64(0)
	start := time.Now()

	// 交替访问 A 和 B 字段
	for i := 0; i < size; i++ {
		if i%2 == 0 {
			sum += slice[i].A
		} else {
			sum += slice[i].B
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("交替访问（[]Fields，A/B 字段交替）: %v, sum=%d\n", elapsed, sum)
	return int64(elapsed)
}

// BenchmarkContiguousInts 测试连续 int64 切片访问的性能
// 使用两个独立的 []int64 切片替代 []Fields
func BenchmarkContiguousInts() int64 {
	const size = 100_000_000
	as := make([]int64, size)
	bs := make([]int64, size)

	// 初始化
	for i := 0; i < size; i++ {
		as[i] = int64(i)
		bs[i] = int64(i)
	}

	sum := int64(0)
	start := time.Now()

	// 顺序访问所有 A（连续内存）
	for i := 0; i < size; i++ {
		sum += as[i]
	}

	elapsed := time.Since(start)
	fmt.Printf("连续访问（两个 []int64，访问 A 切片）: %v, sum=%d\n", elapsed, sum)
	return int64(elapsed)
}

// RunDemo 运行所有演示
func RunDemo() {
	fmt.Println("Cache Line 优化演示")
	fmt.Println("====================")
	fmt.Println("\n场景：访问大量数据时，缓存友好的访问模式性能更优")
	fmt.Println()

	_ = BenchmarkSequentialAccess()
	_ = BenchmarkRandomAccess()
	sum := BenchmarkContiguousInts()

	fmt.Printf("\n最终累加结果: %d\n", sum)
	fmt.Println("\n预期结果分析：")
	fmt.Println("- 顺序访问同一字段（如 BenchmarkSequentialAccess）通常最快")
	fmt.Println("- 交替访问不同字段（如 BenchmarkRandomAccess）可能较慢")
	fmt.Println("- 连续 int64 切片在特定场景下性能最优")
	fmt.Println("\n注意：实际性能差异取决于 CPU 缓存大小和预取策略")
}

func main() {
	RunDemo()
}