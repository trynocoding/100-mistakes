# 内存对齐

## 什么是内存对齐

内存对齐是 CPU 访问内存的一种优化策略。编译器会将结构体字段安排在特定对齐边界上，使 CPU 能够更高效地访问数据。对齐系数通常是字段类型大小或指定的对齐值中最大的那个。

## 结构体字段按对齐系数排列

将字段按对齐系数从大到小排序，可以减少结构体占用的内存空间。

```go
package main

import (
	"fmt"
	"unsafe"
)

// 未优化：字段按声明顺序排列
type Unoptimized struct {
	a bool   // 1 字节，对齐系数 1
	b int64  // 8 字节，对齐系数 8
	c int32  // 4 字节，对齐系数 4
	d string // 16 字节，对齐系数 8
}

// 优化：字段按对齐系数从大到小排列
type Optimized struct {
	b int64  // 8 字节，对齐系数 8
	d string // 16 字节，对齐系数 8
	c int32  // 4 字节，对齐系数 4
	a bool   // 1 字节，对齐系数 1
}

func main() {
	u := Unoptimized{}
	o := Optimized{}

	fmt.Printf("未优化结构体大小: %d 字节\n", unsafe.Sizeof(u))
	fmt.Printf("优化后结构体大小: %d 字节\n", unsafe.Sizeof(o))
}
```

运行结果：

```
未优化结构体大小: 48 字节
优化后结构体大小: 40 字节
```

通过简单地调整字段顺序，节省了 8 字节空间。

## 空结构体作为最后一个字段

空结构体 `struct{}` 不占用空间，但放在结构体末尾时，编译器会为其分配对齐空间，这可能导致结构体数组中后续元素无法正确对齐。

```go
package main

import (
	"fmt"
	"unsafe"
)

type WithEmptyLast struct {
	a int64
	_ struct{} // 空结构体作为最后一个字段
}

type WithoutEmptyLast struct {
	a int64
}

func main() {
	w := WithEmptyLast{}
	wo := WithoutEmptyLast{}

	fmt.Printf("带空结构体末尾: %d 字节\n", unsafe.Sizeof(w))
	fmt.Printf("无空结构体末尾: %d 字节\n", unsafe.Sizeof(wo))
}
```

运行结果：

```
带空结构体末尾: 8 字节
无空结构体末尾: 8 字节
```

虽然单结构体大小相同，但在结构体数组场景下，空结构体末尾可能导致内存浪费：

```go
package main

import (
	"fmt"
	"unsafe"
)

type Element struct {
	Value int32
	_     struct{} // 末尾空结构体影响数组布局
}

func main() {
	arr := []Element{{1}, {2}, {3}}
	fmt.Printf("每个元素大小: %d 字节\n", unsafe.Sizeof(Element{}))
	fmt.Printf("数组总大小: %d 字节\n", unsafe.Sizeof(arr))
}
```

**最佳实践**：避免在结构体末尾放置空结构体字段。如需标记语义，使用 `Padding` 字段代替空结构体。

```go
// 推荐方式：使用具名字段标记为已知的结束填充
type Person struct {
	Age      int32
	Name     string
	EndPad   [0]byte // 明确表示这是结束填充
}
```

## 使用 fieldalignment 工具优化结构体大小

Go 提供了 `fieldalignment` 工具来检测和修复结构体的对齐问题。

### 安装

```bash
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
```

### 使用方法

检测对齐问题：

```bash
fieldalignment -fix ./...
```

### 示例

运行 fieldalignment 前的代码：

```go
type Person struct {
	name string
	age  int
	id   int
}
```

运行 `fieldalignment -fix` 后，自动优化为：

```go
type Person struct {
	age  int
	id   int
	name string
}
```

## 对齐原理详解

### 对齐系数

每种类型都有其对齐系数：

| 类型 | 大小 | 对齐系数 |
|------|------|----------|
| bool | 1 | 1 |
| int8 | 1 | 1 |
| int16 | 2 | 2 |
| int32 | 4 | 4 |
| int64 | 8 | 8 |
| string | 16 | 8 |
| []int | 24 | 8 |
| struct{} | 0 | 1 |

### 填充机制

编译器会在字段之间插入填充（padding），以满足对齐要求。

```go
package main

import (
	"fmt"
	"unsafe"
)

type Demo struct {
	a bool  // 偏移量 0，占用 1 字节
	// 填充 7 字节，使 b 的偏移量达到 8
	b int64 // 偏移量 8，占用 8 字节
	c bool  // 偏移量 16，占用 1 字节
	// 填充 7 字节，使结构体总大小为 24 的倍数
}

func main() {
	d := Demo{}
	fmt.Printf("Demo 大小: %d 字节\n", unsafe.Sizeof(d))
	fmt.Printf("a 偏移量: %d\n", unsafe.Offsetof(d.a))
	fmt.Printf("b 偏移量: %d\n", unsafe.Offsetof(d.b))
	fmt.Printf("c 偏移量: %d\n", unsafe.Offsetof(d.c))
}
```

运行结果：

```
Demo 大小: 24 字节
a 偏移量: 0
b 偏移量: 8
c 偏移量: 16
```

## 总结

1. **按对齐系数排序**：将字段按大小从大到小排序，减少填充
2. **避免空结构体放在末尾**：会占用对齐空间
3. **使用 fieldalignment 工具**：自动化检测和修复对齐问题
4. **理解对齐原理**：有助于编写更高效的数据结构

实际开发中，建议使用 `fieldalignment -fix` 工具来自动优化结构体大小，专注于业务逻辑的实现。
