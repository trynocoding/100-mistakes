package main

import (
	"fmt"
	"io"
	"reflect"
	"sync"
)

// ============================================================================
// 示例 1：嵌入类型意外导出方法的危险
// ============================================================================

// BadConfig 嵌入 sync.Mutex - 危险！
type BadConfig struct {
	sync.Mutex // 嵌入后，Lock/Unlock 方法会被提升到 BadConfig
	values     map[string]string
}

func demonstrateEmbeddingDanger() {
	fmt.Println("=== 嵌入 sync.Mutex 的危险 ===")

	bad := &BadConfig{
		values: make(map[string]string),
	}

	// 问题：Lock/Unlock 被意外导出
	// 任何人都可以直接调用 bad.Lock()

	// 方法被提升到外部类型
	bad.Lock() // 这是 BadConfig 的方法，而不是内部 mu 的方法
	bad.values["db"] = "localhost:5432"
	bad.Unlock()

	// 这意味着 BadConfig 实现了 sync.Locker 接口
	var locker sync.Locker = bad // 完全合法，但可能不是我们的意图
	_ = locker

	fmt.Printf("BadConfig 导出方法: Lock=%v, Unlock=%v\n",
		hasMethod(bad, "Lock"), hasMethod(bad, "Unlock"))
	fmt.Printf("BadConfig 实现了 sync.Locker: %v\n", implementsLocker(bad))
	fmt.Println()
}

// ============================================================================
// 示例 2：正确的做法 - 使用私有字段
// ============================================================================

// GoodConfig 使用私有字段封装 Mutex - 正确做法
type GoodConfig struct {
	mu     sync.Mutex // 私有字段，方法不会被导出
	values map[string]string
}

func (c *GoodConfig) Get(key string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.values[key]
}

func (c *GoodConfig) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

func demonstrateCorrectApproach() {
	fmt.Println("=== 正确的做法：使用私有字段 ===")

	good := &GoodConfig{
		values: make(map[string]string),
	}

	// 正确：通过方法访问
	good.Set("db", "localhost:5432")

	// GoodConfig 不导出 Lock/Unlock
	fmt.Printf("GoodConfig 导出 Lock 方法: %v\n", hasMethod(good, "Lock"))
	fmt.Printf("GoodConfig 实现了 sync.Locker: %v\n", implementsLocker(good))
	fmt.Printf("获取值: %s\n", good.Get("db"))
	fmt.Println()
}

// ============================================================================
// 示例 3：命名冲突问题
// ============================================================================

type Base struct {
	Name string
}

func (b Base) Method() string {
	return "Base.Method: " + b.Name
}

type Derived struct {
	Base
	Name string // 与 Base.Name 同名，会遮蔽基类的字段
}

func (d Derived) Method() string {
	return "Derived.Method: " + d.Name
}

func demonstrateNameConflict() {
	fmt.Println("=== 命名冲突问题 ===")

	d := Derived{
		Base: Base{Name: "BaseName"},
		Name: "DerivedName",
	}

	// Name 字段被遮蔽
	fmt.Printf("d.Name = %q (期望 DerivedName)\n", d.Name)
	fmt.Printf("d.Base.Name = %q (需要通过嵌入访问)\n", d.Base.Name)

	// 方法也会被遮蔽
	fmt.Printf("d.Method() = %q\n", d.Method())
	fmt.Printf("d.Base.Method() = %q (调用基类方法)\n", d.Base.Method())
	fmt.Println()
}

// ============================================================================
// 示例 4：接口意外实现
// ============================================================================

type Reader struct{}

func (Reader) Read(p []byte) (n int, err error) {
	return len(p), nil
}

type Writer struct{}

func (Writer) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// IoWrapper 嵌入两个类型，会自动实现 io.Reader 和 io.Writer
type IoWrapper struct {
	Reader
	Writer
}

func demonstrateInterfacePollution() {
	fmt.Println("=== 接口意外实现 ===")

	// IoWrapper 自动实现了 io.Reader 和 io.Writer
	var _ io.Reader = &IoWrapper{}
	var _ io.Writer = &IoWrapper{}

	// 验证
	wrapper := &IoWrapper{}
	fmt.Printf("IoWrapper 实现了 io.Reader: %v\n", implementsReader(wrapper))
	fmt.Printf("IoWrapper 实现了 io.Writer: %v\n", implementsWriter(wrapper))
	fmt.Println()
}

// ============================================================================
// 示例 5：何时适合使用嵌入
// ============================================================================

// Logger 是可嵌入的基础类型
type Logger struct {
	prefix string
}

func (l Logger) Log(msg string) {
	fmt.Printf("[%s] %s\n", l.prefix, msg)
}

// AppLogger 使用嵌入 - 适合场景
type AppLogger struct {
	Logger          // 嵌入，复用 Logger 的方法
	appName string
}

func (l AppLogger) LogApp(msg string) {
	l.Logger.Log(l.appName + ": " + msg)
}

func demonstrateGoodEmbedding() {
	fmt.Println("=== 适合嵌入的场景 ===")

	appLog := AppLogger{
		Logger:  Logger{prefix: "APP"},
		appName: "MyApp",
	}

	// 复用了 Logger 的 Log 方法
	appLog.Log("系统启动")        // 调用 Logger.Log
	appLog.LogApp("用户登录")     // 自己的方法

	// 嵌入提升了 Logger 的方法
	fmt.Printf("AppLogger 有 Log 方法: %v\n", hasMethod(appLog, "Log"))
	fmt.Println()
}

// ============================================================================
// 辅助函数和类型
// ============================================================================

// hasMethod 检查类型是否有指定方法
func hasMethod(obj any, name string) bool {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	for i := 0; i < objType.NumMethod(); i++ {
		if objType.Method(i).Name == name {
			return true
		}
	}
	return false
}

// implementsLocker 检查类型是否实现了 sync.Locker
func implementsLocker(obj any) bool {
	_, ok := obj.(sync.Locker)
	return ok
}

// implementsReader 检查类型是否实现了 io.Reader
func implementsReader(obj any) bool {
	_, ok := obj.(io.Reader)
	return ok
}

// implementsWriter 检查类型是否实现了 io.Writer
func implementsWriter(obj any) bool {
	_, ok := obj.(io.Writer)
	return ok
}

// main 函数运行所有示例
func main() {
	demonstrateEmbeddingDanger()
	demonstrateCorrectApproach()
	demonstrateNameConflict()
	demonstrateInterfacePollution()
	demonstrateGoodEmbedding()
}
