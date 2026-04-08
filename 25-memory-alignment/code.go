package main

import (
	"fmt"
	"unsafe"
)

// 对齐系数演示

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

// 空结构体演示

type WithEmptyLast struct {
	a int64
	_ struct{} // 空结构体作为最后一个字段
}

type WithoutEmptyLast struct {
	a int64
}

type Element struct {
	Value int32
	_     [0]byte // 末尾填充，使用零长度数组代替空结构体
}

// 对齐原理演示

type Demo struct {
	a bool  // 偏移量 0，占用 1 字节
	// 填充 7 字节，使 b 的偏移量达到 8
	b int64 // 偏移量 8，占用 8 字节
	c bool  // 偏移量 16，占用 1 字节
	// 填充 7 字节，使结构体总大小为 24 的倍数
}

func main() {
	// 1. 结构体字段按对齐系数排列
	fmt.Println("=== 结构体字段按对齐系数排列 ===")
	u := Unoptimized{}
	o := Optimized{}

	fmt.Printf("未优化结构体大小: %d 字节\n", unsafe.Sizeof(u))
	fmt.Printf("优化后结构体大小: %d 字节\n", unsafe.Sizeof(o))
	fmt.Printf("节省空间: %d 字节\n\n", unsafe.Sizeof(u)-unsafe.Sizeof(o))

	// 2. 空结构体作为最后一个字段
	fmt.Println("=== 空结构体作为最后一个字段 ===")
	w := WithEmptyLast{}
	wo := WithoutEmptyLast{}

	fmt.Printf("带空结构体末尾: %d 字节\n", unsafe.Sizeof(w))
	fmt.Printf("无空结构体末尾: %d 字节\n", unsafe.Sizeof(wo))
	fmt.Printf("单元素差异: %d 字节\n\n", unsafe.Sizeof(w)-unsafe.Sizeof(wo))

	// 结构体数组场景
	arr := make([]Element, 3)
	arr[0].Value = 1
	arr[1].Value = 2
	arr[2].Value = 3
	fmt.Printf("Element 每个元素大小: %d 字节\n", unsafe.Sizeof(Element{}))
	fmt.Printf("数组总大小: %d 字节\n\n", unsafe.Sizeof(arr))

	// 3. 对齐原理详解
	fmt.Println("=== 对齐原理详解 ===")
	d := Demo{}
	fmt.Printf("Demo 大小: %d 字节\n", unsafe.Sizeof(d))
	fmt.Printf("a 偏移量: %d\n", unsafe.Offsetof(d.a))
	fmt.Printf("b 偏移量: %d\n", unsafe.Offsetof(d.b))
	fmt.Printf("c 偏移量: %d\n", unsafe.Offsetof(d.c))
}
