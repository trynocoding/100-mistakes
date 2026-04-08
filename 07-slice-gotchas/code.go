package main

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

// 1. 长度与容量的区别
func demonstrateLenCap() {
	fmt.Println("=== 1. 长度与容量的区别 ===")

	s := []int{1, 2, 3, 4, 5}
	fmt.Printf("切片: %v\n", s)
	fmt.Printf("len: %d, cap: %d\n\n", len(s), cap(s))

	// 通过切片创建子切片
	sub := s[1:3]
	fmt.Printf("子切片 s[1:3]: %v\n", sub)
	fmt.Printf("len: %d, cap: %d (从索引1到数组末尾还有4个元素)\n\n", len(sub), cap(sub))

	// 容量解释图示
	fmt.Println("底层数组: [1] [2] [3] [4] [5]")
	fmt.Println("           ^           ^")
	fmt.Println("          sub[0]      sub[cap-1]")
}

// 2. append 扩容机制
func demonstrateAppendGrowth() {
	fmt.Println("\n=== 2. append 扩容机制 ===")

	var s []int

	// 容量 < 1024 时，每次扩容约 2 倍
	fmt.Println("--- 小容量扩容 (< 1024) ---")
	for i := 0; i < 15; i++ {
		s = append(s, i)
		fmt.Printf("len=%d, cap=%d\n", len(s), cap(s))
	}

	fmt.Println("\n--- 大容量扩容 (>= 1024) ---")
	prevCap := cap(s)
	for i := 0; i < 10; i++ {
		s = append(s, 1)
		if cap(s) != prevCap {
			growth := float64(cap(s)-prevCap) / float64(prevCap) * 100
			fmt.Printf("len=%d, cap=%d (增长 %.1f%%)\n", len(s), cap(s), growth)
			prevCap = cap(s)
		}
	}
}

// 3. 切片共享底层数组
func demonstrateSharedArray() {
	fmt.Println("\n=== 3. 切片共享底层数组 ===")

	original := []int{1, 2, 3, 4, 5}
	fmt.Printf("原始切片: %v\n", original)

	sub := original[1:4]
	fmt.Printf("子切片 s[1:4]: %v\n", sub)

	sub[1] = 99
	fmt.Printf("修改 sub[1]=99 后:\n")
	fmt.Printf("  子切片: %v\n", sub)
	fmt.Printf("  原始切片: %v (original[2] 也被修改)\n", original)

	original[1] = 777
	fmt.Printf("修改 original[1]=777 后:\n")
	fmt.Printf("  子切片: %v (sub[0] 也被修改)\n", sub)
	fmt.Printf("  原始切片: %v\n", original)
}

func safeCopy() {
	fmt.Println("\n--- 安全复制方法 ---")

	original := []int{1, 2, 3, 4, 5}

	copy1 := make([]int, len(original))
	copy(copy1, original)

	copy2 := append([]int(nil), original...)

	copy3 := original[:len(original):len(original)]

	original[0] = 999
	fmt.Printf("original: %v\n", original)
	fmt.Printf("copy1 (make+copy): %v\n", copy1)
	fmt.Printf("copy2 (append): %v\n", copy2)
	fmt.Printf("copy3 (三个索引): %v\n", copy3)
}

// 4. nil slice vs empty slice
func demonstrateNilVsEmpty() {
	fmt.Println("\n=== 4. nil slice vs empty slice ===")

	var nilSlice []int
	emptySlice := []int{}
	makeSlice := make([]int, 0)

	fmt.Printf("nilSlice:  len=%d, cap=%d, ptr=%p, isNil=%v\n",
		len(nilSlice), cap(nilSlice), unsafe.Pointer(nil), nilSlice == nil)
	fmt.Printf("emptySlice: len=%d, cap=%d, isNil=%v\n",
		len(emptySlice), cap(emptySlice), emptySlice == nil)
	fmt.Printf("makeSlice:  len=%d, cap=%d, isNil=%v\n",
		len(makeSlice), cap(makeSlice), makeSlice == nil)

	nilSlice = append(nilSlice, 1)
	emptySlice = append(emptySlice, 1)
	makeSlice = append(makeSlice, 1)

	fmt.Println("\nappend 之后:")
	fmt.Printf("nilSlice:  %v\n", nilSlice)
	fmt.Printf("emptySlice: %v\n", emptySlice)
	fmt.Printf("makeSlice:  %v\n", makeSlice)

	nilJSON, _ := json.Marshal(nilSlice)
	emptyJSON, _ := json.Marshal(emptySlice)
	fmt.Printf("\nJSON 序列化:\n")
	fmt.Printf("  nil slice  -> %s\n", nilJSON)
	fmt.Printf("  empty slice -> %s\n", emptyJSON)
}

// 5. copy 注意事项
func demonstrateCopyRequirements() {
	fmt.Println("\n=== 5. copy 注意事项 ===")

	var dstNil []int
	src := []int{1, 2, 3}
	n := copy(dstNil, src)
	fmt.Printf("copy 到 nil slice: n=%d, dst=%v (无效果)\n", n, dstNil)

	dst := make([]int, 3)
	n = copy(dst, src)
	fmt.Printf("copy 到 make slice: n=%d, dst=%v\n", n, dst)

	dst2 := make([]int, 0, 5)
	dst2 = append(dst2, 0, 0, 0, 0, 0)
	n = copy(dst2, src)
	fmt.Printf("copy 到有容量的 slice: n=%d, dst=%v\n", n, dst2)

	dst3 := make([]int, len(src))
	copy(dst3, src)
	fmt.Printf("copy 到正确长度的 slice: n=%d, dst=%v\n", n, dst3)

	fmt.Println("\n--- 部分复制 ---")
	dstPartial := make([]int, 2)
	srcLarge := []int{1, 2, 3, 4}
	n = copy(dstPartial, srcLarge)
	fmt.Printf("目标 len=2, 源 len=4: n=%d, dst=%v\n", n, dstPartial)
}

// 6. capacity 内存泄漏问题
func demonstrateMemoryLeak() {
	fmt.Println("\n=== 6. capacity 内存泄漏问题 ===")

	largeSlice := make([]int, 10, 10000)
	fmt.Printf("largeSlice: len=%d, cap=%d\n", len(largeSlice), cap(largeSlice))

	subset1 := largeSlice[:10]
	fmt.Printf("subset1 (s[:10]): len=%d, cap=%d (容量未变!)\n", len(subset1), cap(subset1))

	subset2 := largeSlice[:10:10]
	fmt.Printf("subset2 (s[:10:10]): len=%d, cap=%d (容量正确)\n", len(subset2), cap(subset2))
}

func preventMemoryLeak() {
	fmt.Println("\n--- 避免内存泄漏的最佳实践 ---")

	big := make([]int, 1000, 10000)
	for i := range big {
		big[i] = i
	}

	subset1 := big[:100]
	subset2 := big[:100:100]
	subset3 := append([]int(nil), big[:100]...)

	fmt.Printf("错误做法 subset1: len=%d, cap=%d\n", len(subset1), cap(subset1))
	fmt.Printf("三个索引 subset2: len=%d, cap=%d\n", len(subset2), cap(subset2))
	fmt.Printf("重新分配 subset3: len=%d, cap=%d\n", len(subset3), cap(subset3))
}

func main() {
	demonstrateLenCap()
	demonstrateAppendGrowth()
	demonstrateSharedArray()
	safeCopy()
	demonstrateNilVsEmpty()
	demonstrateCopyRequirements()
	demonstrateMemoryLeak()
	preventMemoryLeak()
}
