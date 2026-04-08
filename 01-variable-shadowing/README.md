# 变量遮蔽（Variable Shadowing）

## 概念

变量遮蔽是指在嵌套作用域中声明的新变量与外层变量同名，导致外层变量被"遮蔽"的现象。Go 编译器不会报错，但这可能导致意外的 bug。

## 错误示例

### 示例 1：短变量声明遮蔽外层变量

```go
var client *http.Client

if tracing {
    client, err := createClientWithTracing() // 错误：创建了新的局部变量
    if err != nil {
        return err
    }
    // 这里的 client 是新创建的局部变量，外层的 var client 未被修改
}

fmt.Println(client) // 这里是 nil！
```

### 示例 2：循环中的遮蔽

```go
func process(items []string) error {
    var err error
    for _, item := range items {
        err := doSomething(item) // 错误：每次循环都创建新的 err
        if err != nil {
            return err // 返回的是当前迭代的 err，外层的 err 始终为 nil
        }
    }
    return err // 永远返回 nil
}
```

### 示例 3：if 语句中的遮蔽

```go
func getConfig() *Config {
    if cfg := loadConfig(); cfg != nil {
        if cfg := mergeWithDefaults(cfg); cfg != nil { // 遮蔽了外层的 cfg
            return cfg
        }
    }
    return nil // 编译器报错：cfg 未定义
}
```

## 正确写法

### 修复示例 1：使用 `=` 而不是 `:=`

```go
var client *http.Client

if tracing {
    var err error
    client, err = createClientWithTracing() // 正确：赋值给已存在的变量
    if err != nil {
        return err
    }
}
```

### 修复示例 2：明确处理错误

```go
func process(items []string) error {
    for _, item := range items {
        if err := doSomething(item); err != nil { // 在 if 中声明
            return err
        }
    }
    return nil
}
```

### 修复示例 3：使用不同的变量名

```go
func getConfig() *Config {
    cfg := loadConfig()
    if cfg != nil {
        merged := mergeWithDefaults(cfg) // 使用不同的变量名
        if merged != nil {
            return merged
        }
    }
    return nil
}
```

## 验证方法

### 方法 1：使用 `go vet` 配合 shadow 工具

```bash
# 安装 shadow 检查工具
go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest

# 运行检查
go vet -vettool=$(which shadow) ./...
```

### 方法 2：现代 Go 版本的内置检测

Go 1.21+ 内置了对变量遮蔽的检测。在 `if`、`for`、`switch` 语句的初始化中，如果存在变量遮蔽，编译器会发出警告。

```bash
# 启用详细警告
go build -gcflags="-m" ./...

# 检查并显示错误
go build ./... 2>&1 | grep -i shadow
```

### 方法 3：使用 golangci-lint

```bash
# 安装 golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 运行 shadow 检查
golangci-lint run --enable=shadow ./...
```

## 最佳实践

1. **尽量避免嵌套作用域中的同名变量**：使用不同的变量名更清晰
2. **优先使用 `=` 而不是 `:=`**：当需要赋值给已存在的变量时
3. **在 if 语句中声明变量时注意遮蔽**：特别是多个 if 嵌套时
4. **使用 `go vet` 工具链**：在 CI/CD 中集成 shadow 检查
5. **启用 IDE 警告**：大多数 IDE 都能检测变量遮蔽并发出警告

## 常见场景

| 场景 | 错误写法 | 正确写法 |
|------|---------|---------|
| 条件语句中赋值 | `if x := f(); x != nil { ... }` | 使用不同的变量名或在外部声明 |
| 错误声明 | `err := foo(); if err != nil { ... }` | 使用 `=` 或不同的变量名 |
| 循环中的错误处理 | `for { err := foo(); if err != nil { return err } }` | 在 for 外部声明 err |

## 参考

- [Go Wiki: Shadowing](https://github.com/golang/go/wiki/Shadowing)
- [golang.org/x/tools/go/analysis/passes/shadow](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/shadow)
