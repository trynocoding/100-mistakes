# init 函数使用注意事项

## 目录

- [init函数的执行时机](#init函数的执行时机)
- [init函数的缺点](#init函数的缺点)
- [替代方案推荐](#替代方案推荐)
- [代码示例](#代码示例)

---

## init函数的执行时机

Go语言的`init`函数是一个特殊的函数，它在程序启动时自动被调用，具有以下执行顺序规则：

### 执行顺序规则

1. **包级别变量初始化之后**：所有包级别的变量先完成初始化，然后才执行`init`函数
2. **按依赖顺序执行**：如果包A导入了包B，则包B的`init`函数先执行
3. **同一包内按文件名字母顺序执行**：同一个包中的多个`init`函数按文件名字母顺序依次调用
4. **main包最后执行**：所有依赖包初始化完成后，才执行main包的`init`函数

### 执行流程图

```
程序启动
    ↓
包级别变量初始化
    ↓
按依赖顺序执行各包的 init 函数
    ↓
main 包的 init 函数（如果有）
    ↓
main 函数开始执行
```

### 示例

```go
// 包A
package pkgA

import "fmt"

var A = initVar()

func init() {
    fmt.Println("pkgA init 执行")
}

func initVar() int {
    fmt.Println("pkgA 变量初始化")
    return 1
}
```

```go
// main.go
package main

import (
    "fmt"
    _ "02-init-function/pkga" // 匿名导入，触发初始化
)

var M = initMainVar()

func init() {
    fmt.Println("main init 执行")
}

func initMainVar() int {
    fmt.Println("main 变量初始化")
    return 0
}

func main() {
    fmt.Println("main 函数执行")
}
```

---

## init函数的缺点

### 1. 难以测试

**问题**：`init`函数在程序启动时自动执行，这使得单元测试变得困难。

- 测试时无法控制`init`的执行时机
- `init`函数中的逻辑无法单独测试
- 一些初始化逻辑可能依赖外部资源（数据库、文件等），导致测试需要完整的集成环境

```go
// 难以测试的代码示例
package badinit

var cfg *Config

func init() {
    // 从环境变量或配置文件读取配置
    // 测试时很难 mock 这个行为
    cfg = loadConfigFromEnv()
}

func GetConfig() *Config {
    return cfg
}
```

### 2. 增加全局依赖

**问题**：`init`函数隐式地创建全局状态，增加模块间的耦合。

- 难以确定依赖关系：代码审查时很难一眼看出模块间的依赖
- 隐式副作用：`init`函数可能修改全局变量或注册处理器，行为不直观
- 难以追踪问题：当出现问题时，很难定位是哪个`init`函数导致的

### 3. 难以控制执行顺序

**问题**：包的初始化顺序虽然有明确定义，但在实际项目中很难维护。

- 文件重命名可能改变执行顺序
- 重构时容易引入微妙的依赖问题
- 循环依赖会导致编译错误，但逻辑依赖（时序依赖）更难发现

```go
// 潜在问题：时序依赖
package main

var db = initDB() // 依赖配置先加载

func initDB() *DB {
    // 但配置的加载可能在另一个包的 init 中
    return connect(GetConfig().DBURL) // 如果配置还没加载好，这里会出问题
}
```

### 4. 延迟初始化增加内存和启动开销

**问题**：即使某些功能从未使用，`init`函数也会在启动时执行。

- 增加程序的启动时间
- 占用不必要的内存
- 对于需要快速启动的服务影响尤为明显

---

## 替代方案推荐

### 方案一：使用显式初始化函数

**推荐指数：★★★★★**

将初始化逻辑封装到显式的初始化函数中，由调用者决定何时执行。

```go
package config

type Config struct {
    DBURL string
}

var globalCfg *Config

// 显式初始化函数，由 main 或其他入口调用
func Init(cfg *Config) error {
    globalCfg = cfg
    return validateConfig(cfg)
}

func GetConfig() *Config {
    return globalCfg
}
```

**优点**：
- 初始化时机完全可控
- 便于测试时传入 mock 配置
- 错误可以显式传播

### 方案二：使用`sync.Once`延迟初始化

**推荐指数：★★★★☆**

对于确实需要延迟初始化的场景，使用`sync.Once`确保线程安全且只执行一次。

```go
package singleton

import "sync"

var (
    once     sync.Once
    instance *Singleton
)

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
        instance.init() // 初始化逻辑
    })
    return instance
}
```

**优点**：
- 线程安全
- 延迟到首次使用时才初始化
- 可以处理初始化错误（通过 channel 或其他机制）

### 方案三：依赖注入

**推荐指数：★★★★★**

通过依赖注入传递依赖，替代隐式的全局状态。

```go
package service

type Service struct {
    db     *DB
    cache  *Cache
}

func NewService(db *DB, cache *Cache) *Service {
    return &Service{
        db:    db,
        cache: cache,
    }
}

func (s *Service) DoSomething() {
    // 使用注入的依赖
}
```

**优点**：
- 依赖关系显式声明
- 易于测试
- 便于替换实现

### 方案四：使用`wire`或`fx`等依赖注入框架

**推荐指数：★★★☆☆**

对于大型项目，可以使用依赖注入框架来管理初始化。

```go
// 使用 fx 示例
func main() {
    fx.New(
        fx.Provide(NewDB, NewCache, NewService),
        fx.Invoke(Run),
    ).Run()
}
```

### 方案对比表

| 方案 | 可测试性 | 控制度 | 复杂度 | 适用场景 |
|------|---------|--------|--------|---------|
| 显式初始化函数 | 高 | 完全可控 | 低 | 大部分场景 |
| sync.Once | 高 | 首次使用时 | 低 | 单例模式 |
| 依赖注入 | 高 | 完全可控 | 中 | 大型项目 |
| 注入框架 | 高 | 完全可控 | 高 | 超大型项目 |

---

## 代码示例

请参见 [code.go](./code.go) 文件，包含以下示例：

1. **init函数执行顺序演示** - 展示包初始化和init函数的执行时机
2. **init函数的问题演示** - 展示难以测试和全局状态的问题
3. **替代方案演示** - 展示如何使用显式初始化和依赖注入

### 运行代码

```bash
cd 02-init-function
go run code.go
```

---

## 总结

### 为什么现代Go代码应该避免使用init函数

1. **测试性差**：`init`函数中的逻辑难以单独测试，通常需要完整的集成测试
2. **隐式依赖**：增加代码的耦合度，依赖关系不清晰
3. **执行顺序不明确**：虽然Go有明确定义，但维护困难
4. **难以处理错误**：`init`函数不能返回错误，错误处理机制受限
5. **加重启动负担**：即使某些功能不使用，也会增加启动时间和内存占用

### 更好的替代方案

- **显式初始化函数**：最简单直接的方案，推荐作为首选
- **sync.Once**：需要延迟初始化时的选择
- **依赖注入**：大型项目和库的首选方案
- **依赖注入框架**：超大型项目的选择

> **最佳实践**：除非编写的是必须自动初始化的库，否则应该优先使用显式初始化函数。如果必须使用init（例如编写库需要自动注册），应该保持init函数尽可能简单，只做最基本的初始化工作。
