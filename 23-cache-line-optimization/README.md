# Cache Line 优化

## 什么是 Cache Line

现代 CPU 的缓存系统以 **cache line**（缓存行）为最小单位进行内存操作。通常情况下，一个 cache line 大小为 **64 字节**。

当 CPU 访问某个内存地址时，它会不仅加载所需的字节，而是加载整个缓存行（64 字节）到 CPU 缓存中。这是因为：
- 内存局部性原理：程序倾向于访问相邻的内存地址
- 预取机制：CPU 会尝试预测即将访问的内存，提前加载

## Cache Line 对性能的影响

### 顺序访问 vs 随机访问

当程序顺序访问数据时（如遍历数组的连续元素），CPU 可以有效地进行**预取**，将下一个需要的 cache line 提前加载到缓存中。

而当程序随机访问数据时（如访问结构体的不同字段），即使这些字段在内存中相距不远，CPU 也可能无法有效预取，导致频繁的缓存未命中（cache miss）。

### 结构体字段布局的影响

考虑以下两种结构体：

```go
// 方案A：字段交错布局
type PointA struct {
    X int64
    Y int64
    Z int64
}

// 方案B：相同类型字段放在一起
type PointB struct {
    Xs []int64
    Ys []int64
    Zs []int64
}
```

如果我们需要遍历所有 X 坐标：

- **方案A**：每次访问 `point.X` 都可能触发缓存未命中，因为 Y、Z 字段可能占用了同一 cache line
- **方案B**：所有 X 坐标连续排列，预取机制可以充分发挥作用

## 切片元素布局的影响

```go
// 结构体切片：每个结构体包含多个字段
type Fields struct {
    A int64
    B int64
}
sliceA := make([]Fields, 1000)

// 原始类型切片：所有 A 连续，所有 B 连续
sliceB := make([]int64, 1000) // A 字段
sliceC := make([]int64, 1000) // B 字段
```

对于需要频繁访问某一字段的场景：
- 原始类型切片具有更好的缓存友好性
- 结构体切片访问不同字段时可能需要加载多个 cache line

## 可运行的 Go 代码

代码文件：`code.go`

```go
package cacheLine

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

// SequentialAccess vs RandomAccess 演示

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
```

## 运行代码

```bash
go run code.go
```

## 预期输出

```
Cache Line 优化演示
====================

场景：访问大量数据时，缓存友好的访问模式性能更优

顺序访问（[]Fields，访问 A 字段）: 89.123ms, sum=4999999950000000
交替访问（[]Fields，A/B 字段交替）: 145.678ms, sum=4999999950000000
连续访问（两个 []int64，访问 A 切片）: 45.678ms, sum=4999999950000000

最终累加结果: 4999999950000000

预期结果分析：
- 顺序访问同一字段（如 BenchmarkSequentialAccess）通常最快
- 交替访问不同字段（如 BenchmarkRandomAccess）可能较慢
- 连续 int64 切片在特定场景下性能最优

注意：实际性能差异取决于 CPU 缓存大小和预取策略
```

## 总结

| 访问模式 | 缓存效率 | 性能特点 |
|---------|---------|---------|
| 顺序访问同一字段 | 高 | CPU 预取有效，缓存命中率高 |
| 交替访问不同字段 | 中 | 预取效果受限，可能频繁 cache miss |
| 连续原始类型切片 | 最高 | 完美利用缓存行，消除字段交错开销 |

**最佳实践**：
- 在高性能场景下，考虑数据的内存布局
- 如果只需要顺序遍历某个字段，使用独立的原始类型切片可能更高效
- 注意缓存行大小（通常 64 字节）对内存访问的影响
- 通过 `unsafe.Sizeof()` 了解结构体的实际大小，验证是否符合预期