package main

import (
	"fmt"
	"sync"
	"time"
)

// ============ 问题代码：可能导致死锁的 Stringer 实现 ============

// BadUser 是一个有问题的 User 类型
// 它的 String() 方法在持有锁时执行了可能被阻塞的操作
type BadUser struct {
	ID   int
	Name string
	mu   sync.Mutex
}

// String 实现 fmt.Stringer 接口（问题写法）
// 在持有锁时调用 fmt.Sprintf 可能触发间接的 String() 调用
func (u *BadUser) String() string {
	u.mu.Lock()
	defer u.mu.Unlock()

	// 模拟一些需要在锁内完成的操作
	time.Sleep(10 * time.Millisecond)

	// 关键问题：如果这个 String() 是通过 fmt.Errorf 间接调用的，
	// 而 fmt.Errorf 内部也持有锁，就可能形成死锁
	return fmt.Sprintf("BadUser{ID: %d, Name: %s}", u.ID, u.Name)
}

// LogError 模拟在持有锁时记录错误
// 这是真实场景中容易触发死锁的地方
func (u *BadUser) LogError(errMsg string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// 在持有锁时执行一些操作
	time.Sleep(50 * time.Millisecond)

	// 问题写法：在持有锁时调用 fmt.Sprintf
	// 如果 fmt.Sprintf 内部调用 String()（通过 %s 格式化）
	// String() 需要获取 u.mu 锁，但我们已经持有了这个锁
	// 这会导致 String() 阻塞等待锁释放
	// 如果 fmt 内部有另一个 goroutine 持有其他锁并等待 u.mu，
	// 就会形成死锁
	result := fmt.Sprintf("Error for user: %s, message: %s", u.String(), errMsg)
	fmt.Println(result)
}

// ============ 正确代码：分离锁和格式化 ============

// GoodUser 是一个正确实现的 User 类型
type GoodUser struct {
	ID   int
	Name string
	mu   sync.Mutex
}

// String 实现 fmt.Stringer 接口（正确写法）
// 先获取数据，在锁外进行格式化
func (u *GoodUser) String() string {
	// 先获取锁，复制需要的数据
	u.mu.Lock()
	id := u.ID
	name := u.Name
	u.mu.Unlock()

	// 在锁外进行格式化操作
	// 这样永远不会在持有锁时调用 fmt 家族函数
	return fmt.Sprintf("GoodUser{ID: %d, Name: %s}", id, name)
}

// LogError 正确的错误记录方式
func (u *GoodUser) LogError(errMsg string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	// 模拟一些需要在锁内完成的操作
	time.Sleep(50 * time.Millisecond)

	// 正确做法：直接使用已获取的字段值构造字符串
	// 不调用可能阻塞的 fmt 函数（尤其是通过 Stringer 接口的格式化）
	result := fmt.Sprintf("Error for user ID=%d, Name=%s, message: %s", u.ID, u.Name, errMsg)
	fmt.Println(result)
}

// ============ 演示函数 ============

func main() {
	fmt.Println("Stringer 死锁问题演示")
	fmt.Println("=====================")
	fmt.Println()

	// 演示问题写法
	fmt.Println("=== 问题代码演示（BadUser） ===")
	badUser := &BadUser{ID: 1, Name: "Alice"}
	badUser.LogError("validation failed")
	fmt.Println("BadUser 错误已记录（如果没有死锁）")
	fmt.Println()

	// 演示正确写法
	fmt.Println("=== 正确代码演示（GoodUser） ===")
	goodUser := &GoodUser{ID: 2, Name: "Bob"}
	goodUser.LogError("validation failed")
	fmt.Println("GoodUser 错误已记录")
	fmt.Println()

	// 演示 Stringer 接口调用
	fmt.Println("=== Stringer 接口调用 ===")
	fmt.Printf("BadUser String(): %s\n", badUser.String())
	fmt.Printf("GoodUser String(): %s\n", goodUser.String())
	fmt.Println()

	fmt.Println("=== 核心问题总结 ===")
	fmt.Println()
	fmt.Println("死锁的根本原因：")
	fmt.Println("1. fmt.Errorf/Sprintf 内部使用了锁来保证并发安全")
	fmt.Println("2. 当我们在 String() 方法中持有自己的锁，并调用 fmt 函数时")
	fmt.Println("3. 如果 String() 是通过 fmt 间接调用的（格式化时）")
	fmt.Println("4. 就会形成：goroutine A 持有我们的锁等待 fmt 锁，")
	fmt.Println("   goroutine B 持有 fmt 锁等待我们的锁 -> 死锁")
	fmt.Println()
	fmt.Println("真实场景示例：")
	fmt.Println("  // 在 HTTP handler 中")
	fmt.Println("  func (u *User) HandleRequest(w http.ResponseWriter, r *http.Request) {")
	fmt.Println("      u.mu.Lock()")
	fmt.Println("      defer u.mu.Unlock()")
	fmt.Println("      // ... 处理请求 ...")
	fmt.Println("      // 在持有锁时记录错误")
	fmt.Println("      log.Printf('请求处理失败: %s', u.String()) // 可能死锁！")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("解决方案：")
	fmt.Println("1. 不要在持有锁时调用 fmt 家族函数")
	fmt.Println("2. 先复制需要的数据（在锁内），再格式化（在锁外）")
	fmt.Println("3. 使用 fmt.Stringer 时，确保 String() 不会在持有锁时调用 fmt")
	fmt.Println("4. 考虑使用独立的日志字段，不通过 String() 方法")
	fmt.Println()
	fmt.Println("代码示例 - 正确写法：")
	fmt.Println("  func (u *User) String() string {")
	fmt.Println("      u.mu.Lock()")
	fmt.Println("      id, name := u.ID, u.Name  // 先复制数据")
	fmt.Println("      u.mu.Unlock()")
	fmt.Println("      return fmt.Sprintf('User{ID: %d, Name: %s}', id, name)  // 在锁外格式化")
	fmt.Println("  }")
}
