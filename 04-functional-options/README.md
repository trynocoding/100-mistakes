# 函数式选项模式（Functional Options Pattern）

## 什么是函数式选项模式

函数式选项模式是一种在 Go 语言中实现可选参数的惯用模式。它允许使用者通过链式调用来配置对象的行为，而无需暴露大量的构造函数或复杂的配置结构体。

### 核心思想

1. **定义一个 `Option` 类型**，它是一个接受配置目标的函数
2. **为目标结构体实现 `Option` 接口**，通过可变参数接受多个选项
3. **每个选项都是一个返回 `Option` 函数的闭包**

### 为什么使用函数式选项模式

| 传统方式 | 函数式选项模式 |
|---------|---------------|
| 需要多个构造函数 | 一个构造函数 + 多个选项 |
| 不方便设置默认值 | 可以优雅地设置默认值 |
| 扩展困难 | 易于扩展新选项 |
| 不支持链式调用 | 支持链式调用 |

---

## 完整实现示例

### 1. 基本结构

```go
// Server 表示一个服务器配置
type Server struct {
    Host string
    Port int
    Timeout time.Duration
    MaxConn int
}
```

### 2. 定义 Option 类型

```go
// Option 是一个函数类型，用于配置 Server
type Option func(*Server) *Server
```

### 3. 定义可选函数

```go
// WithHost 设置服务器主机地址
func WithHost(host string) Option {
    return func(s *Server) *Server {
        s.Host = host
        return s
    }
}

// WithPort 设置服务器端口
func WithPort(port int) Option {
    return func(s *Server) *Server {
        s.Port = port
        return s
    }
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
    return func(s *Server) *Server {
        s.Timeout = timeout
        return s
    }
}

// WithMaxConn 设置最大连接数
func WithMaxConn(maxConn int) Option {
    return func(s *Server) *Server {
        s.MaxConn = maxConn
        return s
    }
}
```

### 4. 构造函数

```go
// NewServer 创建一个新的服务器配置
// 使用可变参数 options 来接收可选配置
func NewServer(options ...Option) *Server {
    // 设置默认值
    server := &Server{
        Host:     "localhost",
        Port:     8080,
        Timeout:  30 * time.Second,
        MaxConn:  100,
    }

    // 应用所有选项
    for _, option := range options {
        option(server)
    }

    return server
}
```

---

## 错误处理（返回 error）

在实际应用中，我们经常需要在配置过程中进行验证。以下示例展示了如何返回错误：

### 带验证的完整实现

```go
// OptionWithValidate 是一个返回 error 的选项函数类型
type OptionWithValidate func(*Server) error

// validatePort 验证端口号
func validatePort(port int) error {
    if port < 1 || port > 65535 {
        return fmt.Errorf("端口号必须在 1-65535 之间，当前值: %d", port)
    }
    return nil
}

// validateHost 验证主机地址
func validateHost(host string) error {
    if host == "" {
        return errors.New("主机地址不能为空")
    }
    return nil
}

// WithPortValid 创建一个带验证的端口选项
func WithPortValid(port int) OptionWithValidate {
    return func(s *Server) error {
        if err := validatePort(port); err != nil {
            return err
        }
        s.Port = port
        return nil
    }
}

// WithHostValid 创建一个带验证的主机选项
func WithHostValid(host string) OptionWithValidate {
    return func(s *Server) error {
        if err := validateHost(host); err != nil {
            return err
        }
        s.Host = host
        return nil
    }
}
```

### 带错误处理的构造函数

```go
// NewServerWithValidation 创建一个带验证的服务器配置
func NewServerWithValidation(options ...OptionWithValidate) (*Server, error) {
    server := &Server{
        Host:     "localhost",
        Port:     8080,
        Timeout:  30 * time.Second,
        MaxConn:  100,
    }

    // 应用所有选项并收集错误
    for _, option := range options {
        if err := option(server); err != nil {
            return nil, fmt.Errorf("配置服务器失败: %w", err)
        }
    }

    return server, nil
}
```

---

## 使用示例

```go
// 示例 1: 使用默认值
server1 := NewServer()
fmt.Printf("默认配置: %+v\n", server1)

// 示例 2: 使用部分选项
server2 := NewServer(
    WithHost("192.168.1.1"),
    WithPort(9090),
)
fmt.Printf("自定义配置: %+v\n", server2)

// 示例 3: 使用所有选项
server3 := NewServer(
    WithHost("example.com"),
    WithPort(443),
    WithTimeout(60*time.Second),
    WithMaxConn(500),
)
fmt.Printf("完整配置: %+v\n", server3)

// 示例 4: 带验证的配置
server4, err := NewServerWithValidation(
    WithHostValid(""),
)
if err != nil {
    fmt.Printf("验证错误: %v\n", err)
}
```

---

## 最佳实践

1. **提供合理的默认值**: 在构造函数中设置默认值，让调用者只需关注需要自定义的选项

2. **保持选项函数简单**: 每个选项函数应该只完成一个任务，便于组合和测试

3. **使用命名选项**: 通过有意义的函数名（如 `WithHost`、`WithPort`）提高代码可读性

4. **考虑返回 error**: 对于需要验证的选项，使用 `OptionWithValidate` 类型

5. **文档化选项**: 为每个选项函数编写清晰的文档，说明其作用和约束条件

---

## 完整代码

请参见 [code.go](./code.go) 文件，该文件包含所有上述示例的可运行代码。