package main

import (
	"errors"
	"fmt"
	"net/http"
)

// =============================================================================
// 变量遮蔽（Variable Shadowing）示例代码
// Go 1.21+ 可运行
// =============================================================================

// -----------------------------------------------------------------------------
// 示例 1：短变量声明遮蔽外层变量（错误示例）
// -----------------------------------------------------------------------------

// createClientWithTracing 模拟创建一个带追踪功能的 HTTP 客户端
func createClientWithTracing() (*http.Client, error) {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

// shadowExample1 演示变量遮蔽问题
func shadowExample1(tracing bool) (*http.Client, error) {
	var client *http.Client

	if tracing {
		// 错误：:= 创建了新的局部变量 client，外层的 var client 未被修改
		client, err := createClientWithTracing()
		if err != nil {
			return nil, err
		}
		fmt.Printf("内层 client 地址: %p\n", client)
	}

	// 这里的 client 是 nil！
	fmt.Printf("外层 client 地址: %p, 值: %v\n", client, client)
	return client, nil
}

// shadowExample1Fixed 修复后的版本
func shadowExample1Fixed(tracing bool) (*http.Client, error) {
	var client *http.Client
	var err error

	if tracing {
		// 正确：使用 = 赋值给已存在的变量
		client, err = createClientWithTracing()
		if err != nil {
			return nil, err
		}
		fmt.Printf("修复后 client 地址: %p\n", client)
	}

	fmt.Printf("修复后外层 client 地址: %p, 值: %v\n", client, client)
	return client, nil
}

// -----------------------------------------------------------------------------
// 示例 2：循环中的遮蔽（错误示例）
// -----------------------------------------------------------------------------

// doSomething 模拟处理一个 item
func doSomething(item string) error {
	if item == "error" {
		return errors.New("处理失败")
	}
	fmt.Printf("成功处理: %s\n", item)
	return nil
}

// shadowExample2 演示循环中的遮蔽问题
func shadowExample2(items []string) error {
	var err error

	for _, item := range items {
		// 错误：每次迭代都创建新的 err 变量
		err := doSomething(item)
		if err != nil {
			return err
		}
	}

	// 这里的 err 始终是最后一次迭代的值，可能是 nil
	fmt.Printf("循环结束后 err 的值: %v\n", err)
	return err
}

// shadowExample2Fixed 修复后的版本
func shadowExample2Fixed(items []string) error {
	for _, item := range items {
		// 正确：在 if 语句中声明错误变量
		if err := doSomething(item); err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// 示例 3：if 语句中的遮蔽
// -----------------------------------------------------------------------------

// Config 模拟配置结构
type Config struct {
	Name string
}

// loadConfig 模拟加载配置
func loadConfig() *Config {
	return &Config{Name: "default"}
}

// mergeWithDefaults 模拟合并默认配置
func mergeWithDefaults(cfg *Config) *Config {
	return cfg
}

// shadowExample3 演示 if 语句中的遮蔽
func shadowExample3() {
	cfg := loadConfig()
	if cfg != nil {
		// 注意：如果这两个 if 都使用 cfg 变量，会导致编译错误
		// 因为在 Go 1.21+ 中，if 语句初始化中声明的变量会遮蔽外层变量
		if merged := mergeWithDefaults(cfg); merged != nil {
			fmt.Printf("配置名称: %s\n", cfg.Name)
			fmt.Printf("合并后配置: %+v\n", merged)
		}
	}
}

// shadowExample3Fixed 修复后的版本
func shadowExample3Fixed() {
	cfg := loadConfig()
	if cfg != nil {
		merged := mergeWithDefaults(cfg)
		if merged != nil {
			// 正确：使用 merged 变量，或者直接使用 cfg
			fmt.Printf("配置名称: %s\n", cfg.Name)
		}
	}
}

// -----------------------------------------------------------------------------
// 示例 4：switch 语句中的遮蔽
// -----------------------------------------------------------------------------

// handleRequest 模拟处理请求
func handleRequest(method string) error {
	switch method {
	case "GET":
		return nil
	case "POST":
		// 错误：在 case 中声明 err 会遮蔽外层的 err
		if err := errors.New("unsupported"); err != nil {
			return err
		}
	}
	return nil
}

// handleRequestFixed 修复后的版本
func handleRequestFixed(method string) error {
	var err error
	switch method {
	case "GET":
		return nil
	case "POST":
		err = errors.New("unsupported")
		if err != nil {
			return err
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// 主函数：运行所有示例
// -----------------------------------------------------------------------------

func main() {
	fmt.Println("=== 变量遮蔽示例 ===\n")

	// 示例 1
	fmt.Println("--- 示例 1: 短变量声明遮蔽 ---")
	client, err := shadowExample1(true)
	fmt.Printf("返回的 client: %v, err: %v\n\n", client, err)

	fmt.Println("--- 示例 1 修复后 ---")
	clientFixed, errFixed := shadowExample1Fixed(true)
	fmt.Printf("返回的 client: %v, err: %v\n\n", clientFixed, errFixed)

	// 示例 2
	fmt.Println("--- 示例 2: 循环中的遮蔽 ---")
	items := []string{"a", "b", "c"}
	if err := shadowExample2(items); err != nil {
		fmt.Printf("错误: %v\n\n", err)
	}

	fmt.Println("--- 示例 2 修复后 ---")
	if err := shadowExample2Fixed(items); err != nil {
		fmt.Printf("错误: %v\n\n", err)
	}

	// 示例 3
	fmt.Println("--- 示例 3: if 语句中的遮蔽 ---")
	shadowExample3()
	fmt.Println()

	// 示例 4
	fmt.Println("--- 示例 4: switch 中的遮蔽 ---")
	if err := handleRequest("POST"); err != nil {
		fmt.Printf("错误: %v\n\n", err)
	}

	fmt.Println("--- 示例 4 修复后 ---")
	if err := handleRequestFixed("POST"); err != nil {
		fmt.Printf("错误: %v\n\n", err)
	}

	fmt.Println("=== 示例运行完成 ===")
}
