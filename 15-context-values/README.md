# Context Values（上下文值）

## 概念

`context.WithValue` 允许在 Context 中存储和传递请求级别的数据。正确使用 context value 可以实现请求链路追踪、认证信息传递等功能。

## 基本用法

### 存储值

```go
ctx := context.Background()
ctx = context.WithValue(ctx, "userID", 123)
```

### 读取值

```go
userID := ctx.Value("userID") // 返回 interface{}
```

## 关键问题：为什么不能使用 string 作为 key？

使用 string 作为 key 会导致**值覆盖**问题，因为不同的库可能使用相同的 key 名称。

### 错误示例：使用 string 作为 key

```go
// 库 A
ctx = context.WithValue(ctx, "requestID", "req-123")

// 库 B（可能来自不同的团队或包）
ctx = context.WithValue(ctx, "requestID", "req-456") // 覆盖了库 A 的值！
```

当库 B 调用 `ctx.Value("requestID")` 时，得到的是 "req-456" 而不是 "req-123"，库 A 的数据被意外覆盖。

## 正确做法：使用非导出类型作为 key

### 推荐模式：定义自定义 key 类型

```go
type contextKey string

// 定义包级别的 key 变量
var userIDKey contextKey = "userID"
var requestIDKey contextKey = "requestID"

// 使用时
ctx = context.WithValue(ctx, userIDKey, 123)
ctx = context.WithValue(ctx, requestIDKey, "req-123")
```

### 关键点

1. **key 类型是非导出的**（私有类型），外部包无法创建相同的 key
2. **key 值是包级别的常量**，确保同一个 key 总是相同的引用
3. **即使 key 字符串相同，由于类型不同，也不会冲突**

## 完整示例

### 错误实现

```go
func wrongImplementation() {
    ctx := context.Background()

    // 第一层：设置 userID
    ctx = context.WithValue(ctx, "userID", "user-001")
    fmt.Printf("第一层 userID: %v\n", ctx.Value("userID"))

    // 第二层：另一个库也使用 "userID" 作为 key
    ctx = context.WithValue(ctx, "userID", "user-002")
    fmt.Printf("第二层 userID: %v\n", ctx.Value("userID"))

    // 第三层：又一个库覆盖了 "userID"
    ctx = context.WithValue(ctx, "userID", "user-003")
    fmt.Printf("第三层 userID: %v\n", ctx.Value("userID"))

    // 问题：获取第一层的值，发现已经被覆盖了！
    fmt.Printf("回到第一层，userID 应该是 user-001，实际是: %v\n", ctx.Value("userID"))
}
```

### 正确实现

```go
// 定义 key 类型
type key string

// 包级别常量 key
var userIDKey key = "userID"
var requestIDKey key = "requestID"
var traceIDKey key = "traceID"

func correctImplementation() {
    ctx := context.Background()

    // 第一层：使用自定义 key 类型
    ctx = context.WithValue(ctx, userIDKey, "user-001")
    fmt.Printf("第一层 userID: %v\n", ctx.Value(userIDKey))

    // 第二层：另一个库使用自己的 key（即使字符串相同，类型不同则不冲突）
    ctx = context.WithValue(ctx, requestIDKey, "req-456")
    fmt.Printf("第二层 requestID: %v\n", ctx.Value(requestIDKey))
    fmt.Printf("第一层 userID 仍然安全: %v\n", ctx.Value(userIDKey))

    // 第三层：又有一个库使用 traceID
    ctx = context.WithValue(ctx, traceIDKey, "trace-789")
    fmt.Printf("第三层 traceID: %v\n", ctx.Value(traceIDKey))

    // 所有原始值都保持不变
    fmt.Printf("所有值都正确保留 - userID: %v, requestID: %v, traceID: %v\n",
        ctx.Value(userIDKey), ctx.Value(requestIDKey), ctx.Value(traceIDKey))
}
```

## Context Value 的查找机制

Context 的 Value 查找是**沿着 context 链向上查找**的：

```go
ctx := context.Background()
ctx = context.WithValue(ctx, userIDKey, "user-001")
ctx, cancel := context.WithCancel(ctx)
defer cancel()

// 查找 userIDKey 会找到外层 ctx 设置的值
fmt.Println(ctx.Value(userIDKey)) // 输出: user-001
```

## 最佳实践

1. **始终使用自定义 key 类型**：定义 `type key string` 并创建包级别常量

2. **key 应该是 unexported**：防止外部包创建相同的 key

3. **提供类型安全的访问函数**：

```go
type contextKey string

var userIDKey contextKey = "userID"

func GetUserID(ctx context.Context) string {
    if val := ctx.Value(userIDKey); val != nil {
        if str, ok := val.(string); ok {
            return str
        }
    }
    return ""
}
```

4. **只在必要时使用 context value**：对于结构化数据，考虑使用函数参数传递

5. **不要在 context 中存储业务数据**：context value 应该只用于请求级别的元数据

## 常见错误

| 错误写法 | 问题 | 正确写法 |
|---------|------|---------|
| `ctx.Value("userID")` | string key 可能冲突 | `ctx.Value(userIDKey)` |
| 到处定义相同的 string key | 不同包会覆盖值 | 使用 unexported 类型 key |
| 在 context 中存储大量数据 | 增加内存压力 | 只存储必要的元数据 |

## 参考

- [Go Blog: Context and Structs](https://go.dev/blog/context)
- [Go Docs: context.WithValue](https://pkg.go.dev/context#WithValue)
