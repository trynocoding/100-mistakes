package main

import (
	"fmt"
	"sort"
)

// 演示 Range 的常见陷阱

func main() {
	fmt.Println("=== Range 陷阱演示 ===")
	fmt.Println()

	// 陷阱一：值拷贝问题
	fmt.Println("【陷阱一】值拷贝问题")
	{
		nums := []int{1, 2, 3}
		fmt.Printf("遍历前: %v\n", nums)

		// 错误写法：修改的是拷贝，不影响原切片
		for i, n := range nums {
			n = n * 2
			fmt.Printf("  遍历中 nums[%d]=%d, n=%d (修改的是拷贝)\n", i, nums[i], n)
		}
		fmt.Printf("遍历后: %v (未改变)\n\n", nums)

		// 正确写法：使用索引修改原元素
		nums = []int{1, 2, 3}
		for i := range nums {
			nums[i] = nums[i] * 2
		}
		fmt.Printf("正确写法遍历后: %v\n", nums)
	}
	fmt.Println()

	// 陷阱二：指针别名问题
	fmt.Println("【陷阱二】指针别名问题")
	{
		nums := []int{1, 2, 3}

		// 正确写法：始终使用索引获取元素地址（兼容所有 Go 版本）
		var pointers []*int
		for i := range nums {
			pointers = append(pointers, &nums[i])
		}
		fmt.Printf("使用索引获取地址: *pointers[0]=%d, *pointers[1]=%d, *pointers[2]=%d\n",
			*pointers[0], *pointers[1], *pointers[2])

		// 注意：在 Go 1.22+ 中，循环变量改为每轮迭代独立，
		// 因此以下写法在新版本中也是正确的，但使用索引更安全可移植
		fmt.Println("Go 1.22+ 循环变量已独立，但使用索引是最佳实践")
	}
	fmt.Println()

	// 陷阱三：映射遍历顺序随机
	fmt.Println("【陷阱三】映射遍历顺序随机")
	{
		m := map[string]int{"b": 2, "a": 1, "c": 3}

		fmt.Println("直接遍历映射（顺序随机）:")
		for k, v := range m {
			fmt.Printf("  %s=%d\n", k, v)
		}

		// 正确做法：先排序键
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Println("排序后遍历:")
		for _, k := range keys {
			fmt.Printf("  %s=%d\n", k, m[k])
		}
	}
	fmt.Println()

	// 陷阱四：通道遍历
	fmt.Println("【陷阱四】通道遍历")
	{
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		ch <- 3
		close(ch) // 关闭通道

		fmt.Println("使用 range 遍历已关闭的通道:")
		for v := range ch {
			fmt.Printf("  收到值: %d\n", v)
		}
		fmt.Println("通道关闭后遍历自动结束")
	}
	fmt.Println()

	// 综合示例：常见错误与纠正
	fmt.Println("【综合示例】常见错误模式")
	{
		// 场景：需要将切片元素乘以 2 并收集结果
		nums := []int{1, 2, 3}

		// 错误：以为修改了原切片
		result := make([]int, len(nums))
		for i, n := range nums {
			n = n * 2
			result[i] = n
		}
		// 实际上 nums 未被修改
		fmt.Printf("错误写法结果: nums=%v, result=%v\n", nums, result)

		// 正确：直接操作原切片或使用索引
		for i := range nums {
			nums[i] = nums[i] * 2
		}
		fmt.Printf("正确写法结果: nums=%v\n", nums)
	}
}
