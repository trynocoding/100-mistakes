package main

import (
	"fmt"
	"os"
)

func main() {
	// 传统八进制写法的问题
	sum := 100 + 010
	fmt.Printf("传统写法 100 + 010 = %d (期望 110，实际是 108)\n", sum)

	// Go 1.13+ 推荐的 0o 前缀写法
	sum2 := 100 + 0o10
	fmt.Printf("推荐写法 100 + 0o10 = %d\n", sum2)

	// 文件权限示例
	fmt.Println("\n--- 文件权限示例 ---")

	// 创建文件并设置权限
	file, err := os.OpenFile("test_octal.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	file.Close()
	fmt.Println("已创建文件 test_octal.txt，权限为 0o644 (rw-r--r--)")

	// 查看文件信息
	info, err := os.Stat("test_octal.txt")
	if err != nil {
		fmt.Printf("获取文件信息失败: %v\n", err)
		return
	}
	fmt.Printf("文件权限: %o\n", info.Mode().Perm())

	// 清理测试文件
	os.Remove("test_octal.txt")
	fmt.Println("已清理测试文件")

	// 常见权限对比
	fmt.Println("\n--- 常见文件权限 ---")
	fmt.Printf("0755 (rwxr-xr-x): %d\n", 0755)
	fmt.Printf("0644 (rw-r--r--): %d\n", 0644)
	fmt.Printf("0400 (r--------): %d\n", 0400)
	fmt.Printf("0600 (rw-------): %d\n", 0600)
}
