package main

import (
	"errors"
	"fmt"
	"os"
)

// ============ 自定义错误类型 ============

// MyError 是一个自定义错误类型
type MyError struct {
	Code    int
	Message string
	Err     error // 包装底层错误
}

// Error 实现 error 接口
func (e *MyError) Error() string {
	return fmt.Sprintf("错误码 %d: %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *MyError) Unwrap() error {
	return e.Err
}

// NewMyError 创建一个新的 MyError
func NewMyError(code int, msg string) error {
	return &MyError{
		Code:    code,
		Message: msg,
	}
}

// WrapMyError 包装一个底层错误创建 MyError
func WrapMyError(code int, msg string, err error) error {
	return &MyError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

// ============ 模拟资源类型 ============

// ClosableResource 模拟一个需要关闭的资源
type ClosableResource struct {
	name    string
	openErr error
	closeErr error
}

// newClosableResource 创建一个资源
func newClosableResource(name string, openErr, closeErr error) (*ClosableResource, error) {
	if openErr != nil {
		return nil, fmt.Errorf("打开资源 %s 失败: %w", name, openErr)
	}
	return &ClosableResource{
		name:      name,
		closeErr: closeErr,
	}, nil
}

// Close 关闭资源
func (r *ClosableResource) Close() error {
	if r.closeErr != nil {
		return fmt.Errorf("关闭资源 %s 失败: %w", r.name, r.closeErr)
	}
	fmt.Printf("成功关闭资源: %s\n", r.name)
	return nil
}

// ============ 演示函数 ============

// demoWrapUnwrap 演示 Wrap 与 Unwrap
func demoWrapUnwrap() {
	fmt.Println("=== 1. Wrap 与 Unwrap 演示 ===")

	// 创建原始错误
	originalErr := errors.New("原始错误：数据库连接失败")

	// 第一层包装
	wrappedErr1 := fmt.Errorf("执行查询时: %w", originalErr)

	// 第二层包装
	wrappedErr2 := fmt.Errorf("处理用户请求时: %w", wrappedErr1)

	// 第三层包装
	wrappedErr3 := fmt.Errorf("HTTP 处理器时: %w", wrappedErr2)

	fmt.Printf("最终错误: %v\n", wrappedErr3)
	fmt.Println()

	// 使用 errors.Unwrap 解包
	fmt.Println("解包过程:")
	unwrapped1 := errors.Unwrap(wrappedErr3)
	unwrapped2 := errors.Unwrap(unwrapped1)
	unwrapped3 := errors.Unwrap(unwrapped2)
	fmt.Printf("  Unwrap 1 次: %v\n", unwrapped1)
	fmt.Printf("  Unwrap 2 次: %v\n", unwrapped2)
	fmt.Printf("  Unwrap 3 次: %v\n", unwrapped3)
	fmt.Println()

	// 验证错误链
	fmt.Println("错误链验证:")
	fmt.Printf("  errors.Is(wrappedErr3, originalErr): %v\n", errors.Is(wrappedErr3, originalErr))
	fmt.Printf("  errors.Is(wrappedErr2, originalErr): %v\n", errors.Is(wrappedErr2, originalErr))
	fmt.Printf("  errors.Is(wrappedErr1, originalErr): %v\n", errors.Is(wrappedErr1, originalErr))
	fmt.Println()
}

// demoErrorsIs 演示 errors.Is 的使用
func demoErrorsIs() {
	fmt.Println("=== 2. errors.Is 演示 ===")

	// 定义一些 sentinel 错误
	var ErrNotFound = errors.New("资源未找到")
	var ErrPermission = errors.New("权限被拒绝")

	// 创建包装错误
	err := fmt.Errorf("操作失败: %w", ErrNotFound)

	fmt.Printf("错误: %v\n", err)
	fmt.Println()

	// 使用 errors.Is 检查
	fmt.Println("errors.Is 检查:")
	fmt.Printf("  errors.Is(err, ErrNotFound): %v\n", errors.Is(err, ErrNotFound))
	fmt.Printf("  errors.Is(err, ErrPermission): %v\n", errors.Is(err, ErrPermission))
	fmt.Println()

	// 多层包装链
	wrapped1 := fmt.Errorf("层1: %w", err)
	wrapped2 := fmt.Errorf("层2: %w", wrapped1)
	wrapped3 := fmt.Errorf("层3: %w", wrapped2)

	fmt.Printf("多层包装错误: %v\n", wrapped3)
	fmt.Printf("  errors.Is(wrapped3, ErrNotFound): %v\n", errors.Is(wrapped3, ErrNotFound))
	fmt.Println()
}

// demoErrorsAs 演示 errors.As 的使用
func demoErrorsAs() {
	fmt.Println("=== 3. errors.As 演示 ===")

	// 创建自定义错误
	originalErr := NewMyError(404, "用户不存在")

	// 包装错误
	wrappedErr := fmt.Errorf("查询用户时: %w", originalErr)

	fmt.Printf("包装错误: %v\n", wrappedErr)
	fmt.Println()

	// 使用 errors.As 提取
	fmt.Println("errors.As 提取:")
	var myErr *MyError
	if errors.As(wrappedErr, &myErr) {
		fmt.Printf("  成功提取 MyError!\n")
		fmt.Printf("  错误代码: %d\n", myErr.Code)
		fmt.Printf("  错误消息: %s\n", myErr.Message)
	} else {
		fmt.Println("  无法提取 MyError")
	}
	fmt.Println()

	// 提取标准库错误
	fmt.Println("提取标准库错误:")
	pathErr := fmt.Errorf("文件操作: %w", os.ErrNotExist)
	var pathError *os.PathError
	if errors.As(pathErr, &pathError) {
		fmt.Printf("  成功提取 PathError!\n")
		fmt.Printf("  操作: %s\n", pathError.Op)
		fmt.Printf("  路径: %s\n", pathError.Path)
	}
	fmt.Println()
}

// demoDeferCloseWithoutHandling 演示 defer close 错误被忽略的问题
func demoDeferCloseWithoutHandling() {
	fmt.Println("=== 4. Defer Close 错误被忽略（问题演示） ===")

	// 创建一个会失败的资源
	resource, err := newClosableResource("test.txt", nil, errors.New("关闭失败"))
	if err != nil {
		fmt.Printf("创建资源失败: %v\n", err)
		return
	}

	// 错误的做法：defer close 错误被忽略
	defer resource.Close()
	fmt.Println("资源已打开，defer 安排了 Close...")
	fmt.Println("（如果 Close 失败，错误将被忽略）")
	fmt.Println()
}

// demoDeferCloseCorrect 演示正确的 defer close 错误处理方式
func demoDeferCloseCorrect() {
	fmt.Println("=== 5. Defer Close 错误正确处理方式 ===")

	// 方式一：使用闭包捕获错误变量
	fmt.Println("方式一：闭包捕获错误变量")
	resource1, err := newClosableResource("config.json", nil, errors.New("模拟关闭失败"))
	if err != nil {
		fmt.Printf("创建资源失败: %v\n", err)
		return
	}

	defer func() {
		if closeErr := resource1.Close(); closeErr != nil {
			fmt.Printf("  [闭包] 捕获到关闭错误: %v\n", closeErr)
		}
	}()
	fmt.Println("  资源 1 已打开")
	fmt.Println()

	// 方式二：使用命名返回值（推荐）
	fmt.Println("方式二：使用命名返回值 + errors.Join（Go 1.20+）")
}

// doSomethingWithResource 使用命名返回值正确处理 close 错误
func doSomethingWithResource() (err error) {
	resource, err := newClosableResource("data.bin", nil, errors.New("关闭时模拟失败"))
	if err != nil {
		return err
	}

	// 使用 defer 闭包处理关闭错误
	defer func() {
		if closeErr := resource.Close(); closeErr != nil {
			// Go 1.20+ 使用 errors.Join 合并错误
			err = errors.Join(err, closeErr)
		}
	}()

	fmt.Println("  资源已打开，业务逻辑执行中...")
	return nil
}

func demoErrorJoin() {
	fmt.Println("=== 6. errors.Join 合并多个错误（Go 1.20+） ===")

	// 创建一个会有多个错误的场景
	resource, err := newClosableResource("multi.txt", nil, errors.New("关闭失败"))
	if err != nil {
		fmt.Printf("创建资源失败: %v\n", err)
		return
	}

	var combinedErr error
	defer func() {
		if closeErr := resource.Close(); closeErr != nil {
			combinedErr = errors.Join(combinedErr, closeErr)
			fmt.Printf("  合并后的错误: %v\n", combinedErr)
		}
	}()

	// 模拟业务错误
	combinedErr = errors.New("业务处理失败")
	fmt.Printf("  业务错误: %v\n", combinedErr)
	fmt.Println("  资源将在函数返回时关闭并合并错误")
	fmt.Println()
}

// demoCompleteErrorChain 演示完整的错误链处理
func demoCompleteErrorChain() {
	fmt.Println("=== 7. 完整错误链处理示例 ===")

	// 定义应用层错误
	var (
		ErrUserNotFound   = errors.New("用户不存在")
		ErrInvalidInput   = errors.New("无效输入")
		ErrDatabaseError  = errors.New("数据库错误")
	)

	// 模拟服务层错误处理
	processUser := func(userID int) error {
		if userID <= 0 {
			return fmt.Errorf("验证用户ID失败: %w", ErrInvalidInput)
		}

		if userID > 1000 {
			return fmt.Errorf("数据库查询失败: %w", ErrDatabaseError)
		}

		return fmt.Errorf("用户服务查询失败: %w", ErrUserNotFound)
	}

	// 模拟 HTTP 处理层
	handleRequest := func(userID int) error {
		err := processUser(userID)
		if err != nil {
			return fmt.Errorf("处理用户请求失败: %w", err)
		}
		return nil
	}

	testCases := []int{0, 1, 1001}

	for _, userID := range testCases {
		fmt.Printf("测试 userID=%d:\n", userID)

		err := handleRequest(userID)
		if err != nil {
			fmt.Printf("  原始错误: %v\n", err)

			// 检查是否为特定错误类型
			switch {
			case errors.Is(err, ErrInvalidInput):
				fmt.Println("  -> 处理: 无效输入错误")
			case errors.Is(err, ErrDatabaseError):
				fmt.Println("  -> 处理: 数据库错误")
			case errors.Is(err, ErrUserNotFound):
				fmt.Println("  -> 处理: 用户不存在错误")
			default:
				fmt.Println("  -> 处理: 未知错误")
			}

			// 遍历错误链
			fmt.Println("  错误链追踪:")
			chain := []error{err}
			for idx := 0; idx < len(chain); idx++ {
				if idx >= 10 { // 防止无限循环
					break
				}
				unwrapped := errors.Unwrap(chain[idx])
				if unwrapped == nil {
					break
				}
				fmt.Printf("    %d: %v\n", idx+1, unwrapped)
				chain = append(chain, unwrapped)
			}
		}
		fmt.Println()
	}
}

func main() {
	fmt.Println("Go 错误处理完整演示")
	fmt.Println("====================")
	fmt.Println()

	demoWrapUnwrap()
	demoErrorsIs()
	demoErrorsAs()
	demoDeferCloseWithoutHandling()
	demoDeferCloseCorrect()

	// 演示 defer 函数中的错误处理
	fmt.Println("=== 5.1 命名返回值方式演示 ===")
	if err := doSomethingWithResource(); err != nil {
		fmt.Printf("  函数返回错误: %v\n", err)
	}
	fmt.Println()

	demoErrorJoin()
	demoCompleteErrorChain()

	// 演示使用自定义错误类型
	fmt.Println("=== 8. 自定义错误类型完整示例 ===")

	// 创建带底层错误的自定义错误
	underlyingErr := errors.New("底层IO错误")
	customErr := WrapMyError(500, "处理请求失败", underlyingErr)

	fmt.Printf("自定义错误: %v\n", customErr)
	fmt.Printf("  Unwrap 获取底层错误: %v\n", errors.Unwrap(customErr))
	fmt.Println()

	// 在错误链中使用 errors.Is 检查
	if errors.Is(customErr, underlyingErr) {
		fmt.Println("  errors.Is 确认错误链正确")
	}

	// 使用 errors.As 提取
	var myErr *MyError
	if errors.As(customErr, &myErr) {
		fmt.Printf("  errors.As 提取: Code=%d, Message=%s\n", myErr.Code, myErr.Message)
	}
}