//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

// 模拟一个简单的 HTTP 服务器
func startTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Hello, World!"}`))
	}))
}

func main() {
	server := startTestServer()
	defer server.Close()

	fmt.Println("=== HTTP Body 关闭示例 ===")
	fmt.Println()

	// 示例 1: 正确使用 defer resp.Body.Close()
	fmt.Println("1. 正确方式: 使用 defer 关闭 Body")
	example1(server.URL)

	// 示例 2: 读取后显式关闭
	fmt.Println("\n2. 正确方式: 读取后显式 Close()")
	example2(server.URL)

	// 示例 3: 使用 io.Copy 而非 ReadAll
	fmt.Println("\n3. 使用 io.Copy 处理响应体")
	example3(server.URL)

	// 示例 4: Go 1.23+ 自动关闭 (向后兼容演示)
	fmt.Println("\n4. 兼容写法: 显式关闭确保兼容性")
	example4(server.URL)

	fmt.Println("\n=== 所有示例执行完成 ===")
}

// example1 演示正确的 defer 关闭方式
func example1(url string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("   请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close() // 正确：在函数返回前关闭

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   读取失败: %v\n", err)
		return
	}
	fmt.Printf("   状态码: %d\n", resp.StatusCode)
	fmt.Printf("   响应内容: %s\n", string(body))
	fmt.Println("   Body 已通过 defer 关闭")
}

// example2 演示读取后显式关闭
func example2(url string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("   请求失败: %v\n", err)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close() // 失败时也要关闭
		fmt.Printf("   读取失败: %v\n", err)
		return
	}
	resp.Body.Close() // 读取成功后显式关闭

	fmt.Printf("   状态码: %d\n", resp.StatusCode)
	fmt.Printf("   响应内容: %s\n", string(body))
	fmt.Println("   Body 已显式关闭")
}

// example3 演示使用 io.Copy 处理响应体
func example3(url string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("   请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// io.Copy 自动处理数据流动，但 Body 仍需关闭
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   读取失败: %v\n", err)
		return
	}
	fmt.Printf("   状态码: %d\n", resp.StatusCode)
	fmt.Printf("   响应内容: %s\n", string(body))
	fmt.Println("   注意：即使使用 io.Copy，Body 仍需关闭")
}

// example4 演示推荐的兼容写法
func example4(url string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("   请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close() // Go 1.23+ 也会自动关闭，但显式关闭是最佳实践

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   读取失败: %v\n", err)
		return
	}
	fmt.Printf("   状态码: %d\n", resp.StatusCode)
	fmt.Printf("   响应内容: %s\n", string(body))
	fmt.Println("   推荐：始终显式关闭，确保跨版本兼容")
}