# Go 错误处理（Error Handling）

## 概述

Go 语言使用显式的错误处理模式，这与许多传统语言的异常处理机制不同。理解和正确使用 Go 的错误处理是编写健壮 Go 代码的基础。

### 错误处理的核心原则

1. **错误是值**: 在 Go 中，错误是一个普通的接口类型 `error`
2. **显式处理**: 每个可能返回错误的函数调用都必须被处理
3. **不要忽略错误**: 使用 `_` 忽略错误通常是一个代码异味的信号

---

## Wrap 与 Unwrap

### 什么是 Wrap

Wrap 是指将一个错误包装在另一个包含更多上下文的错误中，同时保留原始错误的信息。这对于调试和问题定位至关重要。

### 使用 `%w` 进行 Wrap

`fmt.Errorf` 使用 `%w` 格式化 verb 可以创建一个包装错误：

```go
if err != nil {
    return fmt.Errorf("操作失败: %w", err)
}
```

### Unwrap 的作用

被包装的错误可以通过 `errors.Unwrap()` 函数取出：

```go
wrappedErr := fmt.Errorf("上层错误: %w", originalErr)
unwrappedErr := errors.Unwrap(wrappedErr)
// unwrappedErr == originalErr
```

### 完整错误链

错误可以多层嵌套：

```go
err1 := errors.New("原始错误")

err2 := fmt.Errorf("中间层: %w", err1)

err3 := fmt.Errorf("最外层: %w", err2)

// err3 -> err2 -> err1
```

---

## errors.Is 与 errors.As

### errors.Is

`errors.Is` 用于检查错误链中是否存在某个特定的错误。它会遍历整个错误链进行匹配。

**使用场景**: 当你需要检查错误是否为某种类型时使用。

```go
// 检查是否包含某个特定错误
if errors.Is(err, os.ErrNotExist) {
    // 文件不存在处理
}
```

### errors.As

`errors.As` 用于从错误链中提取特定类型的错误。它会找到第一个匹配类型的错误并将其赋值给目标变量。

**使用场景**: 当你需要获取错误的详细信息时使用。

```go
// 提取自定义错误类型
var myErr *MyError
if errors.As(err, &myErr) {
    fmt.Printf("错误代码: %d\n", myErr.Code)
}
```

### 区别对比

| 特性 | errors.Is | errors.As |
|------|----------|----------|
| 用途 | 检查错误是否存在 | 提取错误并赋值 |
| 返回值 | bool | bool |
| 参数 | 目标错误值 | 目标错误指针 |
| 修改原变量 | 否 | 是（通过指针） |

---

## Defer 中处理 Close 错误

### 问题背景

在 defer 中直接调用 Close 方法时，如果 Close 返回错误，这个错误很可能会被忽略：

```go
// 错误示例
file, _ := os.Open("test.txt")
defer file.Close() // 如果 Close 失败，错误被丢弃
```

### 正确方式一：闭包捕获错误变量

```go
file, err := os.Open("test.txt")
if err != nil {
    return err
}
defer func() {
    if closeErr := file.Close(); closeErr != nil {
        // 记录或处理 close 错误
        fmt.Printf("关闭文件失败: %v\n", closeErr)
    }
}()
```

### 正确方式二：使用命名的返回值

```go
func doSomething() (err error) {
    file, err := os.Open("test.txt")
    if err != nil {
        return err
    }
    defer func() {
        if closeErr := file.Close(); closeErr != nil {
            // 使用 errors.Join 合并错误（Go 1.20+）
            err = errors.Join(err, closeErr)
        }
    }()
    // 业务逻辑
    return nil
}
```

### Go 1.20+ 的 errors.Join

从 Go 1.20 开始，可以使用 `errors.Join` 合并多个错误：

```go
err := errors.Join(err1, err2, err3)
```

---

## 自定义错误类型

### 使用 struct 实现 error 接口

```go
type MyError struct {
    Code    int
    Message string
    Err     error // 包装底层错误
}

func (e *MyError) Error() string {
    return fmt.Sprintf("错误码 %d: %s", e.Code, e.Message)
}

func (e *MyError) Unwrap() error {
    return e.Err
}
```

### 创建自定义错误的工厂函数

```go
func NewMyError(code int, msg string) error {
    return &MyError{
        Code:    code,
        Message: msg,
    }
}

func WrapMyError(code int, msg string, err error) error {
    return &MyError{
        Code:    code,
        Message: msg,
        Err:     err,
    }
}
```

---

## 最佳实践

1. **始终包装错误**: 使用 `%w` 为错误添加上下文信息

2. **不要两次返回同一错误**: 如果要包装错误，不要同时返回原始错误

3. **使用 errors.Is/As**: 在检查包装错误时使用标准库方法，而非直接比较

4. **Defer 中处理 Close 错误**: 始终在 defer 中正确处理可能失败的 Close 调用

5. **错误只处理一次**: 要么记录错误并继续，要么返回给调用者，不要同时做两件事

6. **使用有意义的错误消息**: 错误消息应该包含足够的上下文信息

---

## 完整代码

请参见 [code.go](./code.go) 文件，该文件包含所有上述示例的可运行代码。