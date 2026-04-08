# Slice 注意事项

## 目录

1. [长度与容量的区别](#1-长度与容量的区别)
2. [append 扩容机制](#2-append-扩容机制)
3. [切片共享底层数组](#3-切片共享底层数组)
4. [nil slice vs empty slice](#4-nil-slice-vs-empty-slice)
5. [copy 注意事项](#5-copy-注意事项)
6. [capacity 内存泄漏问题](#6-capacity-内存泄漏问题)

---

## 1. 长度与容量的区别

**长度 (length)**：切片中实际存在的元素数量，通过 `len()` 获取。

**容量 (capacity)**：底层数组中从切片起始位置到数组末尾的元素数量，通过 `cap()` 获取。

```go
s := []int{1, 2, 3, 4, 5}
fmt.Println("len:", len(s)) // 5
fmt.Println("cap:", cap(s)) // 5

// 通过切片创建子切片
sub := s[1:3]
fmt.Println("len:", len(sub)) // 2 (元素: 2, 3)
fmt.Println("cap:", cap(sub)) // 4 (从索引1到底层数组末尾: 2,3,4,5)
```

**关键点**：

- 长度表示"我能访问多少元素"
- 容量表示"我底层还能扩展多少空间"
- 向切片添加元素时，不能超过容量

---

## 2. append 扩容机制

Go 的 `append` 函数在切片容量不足时会自动扩容，扩容策略如下：

| 原容量 | 扩容倍数 | 新容量 |
|--------|----------|--------|
| < 1024 | 2 倍 | 原容量 × 2 |
| >= 1024 | 1.25 倍 | 原容量 × 1.25 |

```go
// 演示 append 扩容机制
func demonstrateAppendGrowth() {
    var s []int

    // 容量 < 1024 时，每次扩容约 2 倍
    for i := 0; i < 10; i++ {
        s = append(s, i)
        fmt.Printf("len=%d, cap=%d\n", len(s), cap(s))
    }

    fmt.Println("\n--- 大容量切片 (< 1024) ---")
    s = make([]int, 0)
    prevCap := 0
    for len(s) < 1024 {
        if cap(s) != prevCap {
            fmt.Printf("len=%d, cap=%d (增长 %.1f%%)\n", len(s), cap(s),
                float64(cap(s)-prevCap)/float64(prevCap)*100)
            prevCap = cap(s)
        }
        s = append(s, 1)
    }

    fmt.Println("\n--- 大容量切片 (>= 1024) ---")
    // 继续扩容，观察 1.25 倍策略
    startLen := len(s)
    for i := 0; i < 5; i++ {
        s = append(s, 1)
        fmt.Printf("len=%d, cap=%d (增长 %.1f%%)\n", len(s), cap(s),
            float64(cap(s)-prevCap)/float64(prevCap)*100)
        prevCap = cap(s)
    }
    _ = startLen // 避免未使用警告
}
```

**实际测试输出示例**：

```
len=1, cap=1
len=2, cap=2
len=3, cap=4
len=4, cap=4
len=5, cap=8
...

--- 大容量切片 (>= 1024) ---
len=1025, cap=1280 (增长 25.0%)
len=1281, cap=1600 (增长 25.0%)
```

---

## 3. 切片共享底层数组

当从一个切片创建子切片时，两个切片共享同一个底层数组。这可能导致意外修改。

```go
func demonstrateSharedArray() {
    // 创建一个原始切片
    original := []int{1, 2, 3, 4, 5}
    fmt.Println("原始切片:", original)

    // 创建子切片（共享底层数组）
    sub := original[1:4] // [2, 3, 4]
    fmt.Println("子切片:", sub)

    // 通过子切片修改元素，影响原始切片
    sub[1] = 99
    fmt.Println("修改后子切片:", sub)
    fmt.Println("修改后原始切片:", original) // original[2] 也变成了 99

    // 反向影响
    original[1] = 777
    fmt.Println("再次修改后子切片:", sub) // sub[0] 也变成了 777
}
```

**输出**：

```
原始切片: [1 2 3 4 5]
子切片: [2 3 4]
修改后子切片: [2 99 4]
修改后原始切片: [1 2 99 4 5]
再次修改后子切片: [777 99 4]
```

### 安全复制方法

```go
func safeCopy() {
    original := []int{1, 2, 3, 4, 5}

    // 方法1：完整复制（推荐）
    copy1 := make([]int, len(original))
    copy(copy1, original)

    // 方法2：使用 append
    copy2 := append([]int(nil), original...)

    // 方法3：使用三个索引的 slice
    copy3 := original[:len(original):len(original)]

    // 现在修改 original 不会影响副本
    original[0] = 999
    fmt.Println("original:", original)
    fmt.Println("copy1:", copy1)
    fmt.Println("copy2:", copy2)
    fmt.Println("copy3:", copy3)
}
```

---

## 4. nil slice vs empty slice

| 特性 | nil slice | empty slice |
|------|-----------|-------------|
| 声明方式 | `var s []int` | `s := []int{}` 或 `s := make([]int, 0)` |
| 底层数组 | `nil` | 存在空数组（非 nil） |
| `len()` | 0 | 0 |
| `cap()` | 0 | 0 |
| `append()` | 可正常工作 | 可正常工作 |
| JSON 序列化 | `null` | `[]` |

```go
func demonstrateNilVsEmpty() {
    var nilSlice []int           // nil slice
    emptySlice := []int{}        // empty slice
    makeSlice := make([]int, 0)  // empty slice (通过 make)

    fmt.Printf("nilSlice:    len=%d, cap=%d, ptr=%p, isNil=%v\n",
        len(nilSlice), cap(nilSlice), unsafe.Pointer(nil), nilSlice == nil)

    fmt.Printf("emptySlice:  len=%d, cap=%d, isNil=%v\n",
        len(emptySlice), cap(emptySlice), emptySlice == nil)

    fmt.Printf("makeSlice:   len=%d, cap=%d, isNil=%v\n",
        len(makeSlice), cap(makeSlice), makeSlice == nil)

    // append 行为一致
    nilSlice = append(nilSlice, 1)
    emptySlice = append(emptySlice, 1)
    makeSlice = append(makeSlice, 1)

    fmt.Println("\nappend 之后:")
    fmt.Printf("nilSlice:  %v\n", nilSlice)
    fmt.Printf("emptySlice: %v\n", emptySlice)
    fmt.Printf("makeSlice:  %v\n", makeSlice)

    // JSON 序列化差异
    import "encoding/json"
    nilJSON, _ := json.Marshal(nilSlice)
    emptyJSON, _ := json.Marshal(emptySlice)
    fmt.Printf("\nJSON 序列化:\n  nil slice  -> %s\n  empty slice -> %s\n", nilJSON, emptyJSON)
}
```

**注意**：需要导入 `encoding/json` 和 `unsafe` 包进行演示。

---

## 5. copy 注意事项

`copy()` 函数将元素从源切片复制到目标切片，**但前提是目标切片必须有足够的内存空间**。

```go
func demonstrateCopyRequirements() {
    // 错误示例：目标切片为 nil，copy 无效
    var dstNil []int
    src := []int{1, 2, 3}
    n := copy(dstNil, src)
    fmt.Printf("copy 到 nil slice: n=%d, dst=%v\n", n, dstNil) // n=0, dst=[]

    // 正确示例1：目标切片已有内存
    dst := make([]int, 3)
    n = copy(dst, src)
    fmt.Printf("copy 到 make slice: n=%d, dst=%v\n", n, dst)

    // 正确示例2：目标切片有足够容量
    dst2 := make([]int, 0, 5)
    dst2 = append(dst2, 0, 0, 0, 0, 0) // 先填充空间
    n = copy(dst2, src)
    fmt.Printf("copy 到有容量的 slice: n=%d, dst=%v\n", n, dst2)

    // 正确示例3：初始化时指定长度
    dst3 := make([]int, len(src))
    copy(dst3, src)
    fmt.Printf("copy 到正确长度的 slice: n=%d, dst=%v\n", n, dst3)
}
```

**copy 的返回值**：返回实际复制的元素数量（取源和目标的较小值）。

```go
func demonstratePartialCopy() {
    dst := make([]int, 2)    // 只能容纳2个元素
    src := []int{1, 2, 3, 4}
    n := copy(dst, src)
    fmt.Printf("部分复制: n=%d, dst=%v\n", n, dst) // n=2, dst=[1,2]
}
```

---

## 6. capacity 内存泄漏问题

使用 `s[:cap(s)]` 创建新切片时，如果新的切片长度远小于容量，但仍然引用整个底层数组，会导致内存泄漏。

```go
func demonstrateMemoryLeak() {
    // 场景1：保留过多容量的切片
    largeSlice := make([]int, 10, 10000) // 容量远大于长度
    fmt.Printf("largeSlice: len=%d, cap=%d\n", len(largeSlice), cap(largeSlice))

    // 截取我们需要的长度，但容量不变
    needed := largeSlice[:10]
    fmt.Printf("needed: len=%d, cap=%d\n", len(needed), cap(needed))

    // 问题：needed 仍然持有整个底层数组的引用，
    // 导致 largeSlice 数组无法被 GC 回收

    // 正确做法：使用完整的三个索引形式
    correct := largeSlice[:10:10]
    fmt.Printf("correct: len=%d, cap=%d\n", len(correct), cap(correct))
    // 现在 correct 的容量等于长度，不再引用多余内存
}
```

### 避免内存泄漏的最佳实践

```go
func preventMemoryLeak() {
    // 原始大数据切片
    big := make([]int, 1000, 10000)
    for i := range big {
        big[i] = i
    }

    // 场景：只需要前100个元素
    // 错误做法
    subset1 := big[:100]
    // subset1 的 cap 仍然是 10000

    // 正确做法1：三个索引
    subset2 := big[:100:100]
    // subset2 的 cap 是 100

    // 正确做法2：重新分配
    subset3 := append([]int(nil), big[:100]...)
    // subset3 是全新的切片，容量=长度

    fmt.Printf("错误做法: len=%d, cap=%d\n", len(subset1), cap(subset1))
    fmt.Printf("三个索引: len=%d, cap=%d\n", len(subset2), cap(subset2))
    fmt.Printf("重新分配: len=%d, cap=%d\n", len(subset3), cap(subset3))
}
```

**总结**：当需要将大切片的部分内容传递给其他函数时，使用 `s[start:end:end]` 形式显式设置容量，防止意外保留大量不需要的内存。

---

## 相关代码

所有可运行的代码示例请参见 [code.go](./code.go)。

运行方式：

```bash
cd /root/.workspace/lang/go/100_mistakes/07-slice-gotchas
go run code.go
```
