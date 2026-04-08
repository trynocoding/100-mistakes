// 代码示例：init函数使用注意事项
// 运行方式：go run code.go
//
// 本文件包含以下示例：
// 1. init函数执行顺序演示
// 2. init函数的问题演示（难以测试）
// 3. 替代方案：显式初始化
// 4. 替代方案：sync.Once延迟初始化
// 5. 替代方案：依赖注入

package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// 示例1：init函数执行顺序演示
// ============================================================================

// 通过两个包来演示执行顺序
// 由于Go不允许在一个文件中创建多个包，我们使用注释说明顺序
//
// 假设有以下文件结构：
//   - pkga/init_demo.go
//   - pkgb/depend.go
//   - main.go
//
// pkga/init_demo.go:
//   package pkga
//   import "fmt"
//   var A = pa()
//   func init() { fmt.Println("1. pkga init 执行") }
//   func pa() int { fmt.Println("1. pkga 变量初始化"); return 1 }
//
// pkgb/depend.go:
//   package pkgb
//   import "fmt"
//   import "02-init-function/pkga"
//   var B = pb()
//   func init() { fmt.Println("2. pkgb init 执行（依赖 pkga）") }
//   func pb() int { fmt.Println("2. pkgb 变量初始化，使用 pkga.A =", pkga.A); return 2 }
//
// main.go:
//   package main
//   import (
//       "fmt"
//       _ "02-init-function/pkga"  // 匿名导入，只为触发初始化
//       _ "02-init-function/pkgb"
//   )
//   var M = pm()
//   func init() { fmt.Println("3. main init 执行") }
//   func pm() int { fmt.Println("3. main 变量初始化"); return 0 }
//   func main() { fmt.Println("4. main 函数执行") }
//
// 输出顺序：
//   1. pkga 变量初始化
//   1. pkga init 执行
//   2. pkgb 变量初始化，使用 pkga.A = 1
//   2. pkgb init 执行（依赖 pkga）
//   3. main 变量初始化
//   3. main init 执行
//   4. main 函数执行

// ============================================================================
// 示例2：init函数的问题演示
// ============================================================================

// 问题1：难以测试 - 全局状态在init中初始化
var globalConfig *Config

type Config struct {
	DBURL      string
	CacheSize  int
	EnableLogs bool
}

// init函数中的初始化逻辑难以单独测试
// 这是BAD PRACTICE - 演示不推荐的做法
func init() {
	// 模拟从环境变量加载配置
	globalConfig = &Config{
		DBURL:      getEnvOrDefault("DB_URL", "localhost:5432"),
		CacheSize:  100,
		EnableLogs: true,
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	// 简化实现，实际应使用os.Getenv
	return defaultVal
}

func GetGlobalConfig() *Config {
	return globalConfig
}

// 问题2：隐式依赖 - 不知道cfg依赖什么
var cfg = loadConfig()

func loadConfig() *Config {
	// 隐式依赖，无法控制加载时机
	time.Sleep(10 * time.Millisecond) // 模拟延迟
	return &Config{DBURL: "implicit-config"}
}

func GetImplicitConfig() *Config {
	return cfg
}

// ============================================================================
// 示例3：替代方案 - 显式初始化函数（推荐）
// ============================================================================

// 使用显式初始化的服务
type DatabaseService struct {
	connection string
	connected  bool
}

// NewDatabaseService 创建数据库服务（需要显式调用Init）
func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

// Init 显式初始化方法 - 由调用者决定何时执行
func (d *DatabaseService) Init(connection string) error {
	d.connection = connection
	// 模拟连接
	time.Sleep(10 * time.Millisecond)
	d.connected = true
	return nil
}

func (d *DatabaseService) Query(sql string) error {
	if !d.connected {
		return fmt.Errorf("not connected, call Init first")
	}
	fmt.Printf("Executing SQL on %s: %s\n", d.connection, sql)
	return nil
}

func (d *DatabaseService) IsConnected() bool {
	return d.connected
}

// ============================================================================
// 示例4：替代方案 - sync.Once 延迟初始化
// ============================================================================

type LazyService struct {
	once     sync.Once
	data     string
	instance *LazyService
}

func NewLazyService() *LazyService {
	return &LazyService{}
}

// GetInstance 使用sync.Once实现线程安全的延迟初始化
func (s *LazyService) GetInstance() *LazyService {
	s.once.Do(func() {
		// 这个函数只会被执行一次
		time.Sleep(10 * time.Millisecond) // 模拟耗时初始化
		s.data = "Lazy initialized data"
		s.instance = s
		fmt.Println("LazyService: 初始化完成（只执行一次）")
	})
	return s.instance
}

func (s *LazyService) GetData() string {
	return s.data
}

// ============================================================================
// 示例5：替代方案 - 依赖注入
// ============================================================================

// 接口定义
type Logger interface {
	Log(msg string)
}

type Cache interface {
	Get(key string) (string, bool)
	Set(key string, value string)
}

// 业务服务使用依赖注入
type UserService struct {
	db    *DatabaseService
	cache Cache
	log   Logger
}

// NewUserService 依赖注入版本的构造函数
func NewUserService(db *DatabaseService, cache Cache, log Logger) *UserService {
	return &UserService{
		db:    db,
		cache: cache,
		log:   log,
	}
}

func (s *UserService) GetUser(id string) error {
	// 1. 先查缓存
	if user, ok := s.cache.Get(id); ok {
		s.log.Log(fmt.Sprintf("Cache hit for user %s: %s", id, user))
		return nil
	}

	// 2. 缓存未命中，查询数据库
	s.log.Log(fmt.Sprintf("Cache miss for user %s, querying DB", id))
	return s.db.Query("SELECT * FROM users WHERE id = " + id)
}

// 简单的日志实现
type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(msg string) {
	fmt.Printf("[LOG] %s\n", msg)
}

// 简单的内存缓存实现
type MemoryCache struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{data: make(map[string]string)}
}

func (m *MemoryCache) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

func (m *MemoryCache) Set(key string, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// ============================================================================
// 主函数演示
// ============================================================================

func main() {
	fmt.Println("========================================")
	fmt.Println("Go init 函数注意事项 - 代码演示")
	fmt.Println("========================================")

	// 演示1：init函数的问题
	fmt.Println("\n--- 演示1：init函数导致的全局状态问题 ---")
	fmt.Printf("全局配置已通过init自动加载: DBURL=%s\n", GetGlobalConfig().DBURL)
	fmt.Printf("隐式配置已加载: DBURL=%s\n", GetImplicitConfig().DBURL)

	// 演示2：显式初始化（推荐）
	fmt.Println("\n--- 演示2：显式初始化（推荐方式）---")
	db := NewDatabaseService()
	fmt.Println("DatabaseService 创建完成，connected:", db.IsConnected())

	// 显式调用Init - 初始化时机完全可控
	err := db.Init("postgres://localhost:5432/mydb")
	if err != nil {
		fmt.Println("初始化失败:", err)
	} else {
		fmt.Println("DatabaseService 初始化成功，connected:", db.IsConnected())
		db.Query("SELECT * FROM users")
	}

	// 演示3：sync.Once延迟初始化
	fmt.Println("\n--- 演示3：sync.Once 延迟初始化 ---")
	lazy := NewLazyService()
	fmt.Println("首次调用 GetInstance():")
	s1 := lazy.GetInstance()
	fmt.Printf("  获取到的数据: %s\n", s1.GetData())

	fmt.Println("第二次调用 GetInstance():")
	s2 := lazy.GetInstance()
	fmt.Printf("  获取到的数据: %s\n", s2.GetData())

	fmt.Println("两次调用返回同一实例:", s1 == s2)

	// 演示4：依赖注入
	fmt.Println("\n--- 演示4：依赖注入（测试友好）---")
	logger := &ConsoleLogger{}
	cache := NewMemoryCache()
	cache.Set("user123", "Alice")

	// 通过构造函数注入依赖
	userService := NewUserService(db, cache, logger)

	// 第一次查询 - 缓存未命中
	fmt.Println("查询用户123（首次，缓存未命中）:")
	userService.GetUser("user123")

	// 设置缓存
	cache.Set("user456", "Bob")

	// 第二次查询 - 缓存命中
	fmt.Println("\n查询用户456（已缓存）:")
	userService.GetUser("user456")

	// 演示5：对比总结
	fmt.Println("\n========================================")
	fmt.Println("总结：为什么应该避免使用 init 函数")
	fmt.Println("========================================")
	fmt.Println("1. 测试困难：全局状态在init中初始化，无法单独测试")
	fmt.Println("2. 隐式依赖：依赖关系不清晰，难以维护")
	fmt.Println("3. 执行顺序：init执行顺序难以控制，容易引入bug")
	fmt.Println("4. 错误处理：init无法返回错误，错误处理受限")
	fmt.Println("\n推荐做法：")
	fmt.Println("  - 使用显式初始化函数")
	fmt.Println("  - 使用 sync.Once 进行延迟初始化")
	fmt.Println("  - 使用依赖注入管理依赖")
}
