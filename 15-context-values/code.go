package main

import (
	"context"
	"fmt"
)

// =============================================================================
// Context Values（上下文值）示例代码
// Go 1.21+ 可运行
// =============================================================================

// -----------------------------------------------------------------------------
// 定义自定义 key 类型（正确做法）
// -----------------------------------------------------------------------------

// key 是非导出类型，用于避免 context value key 冲突
type key string

// 包级别的 key 常量
var (
	userIDKey    key = "userID"
	requestIDKey key = "requestID"
	traceIDKey   key = "traceID"
)

// -----------------------------------------------------------------------------
// 错误示例：使用 string 作为 key 导致值被覆盖
// -----------------------------------------------------------------------------

// wrongUserIDKey 模拟使用 string 作为 key
func wrongUserIDKey() string {
	return "userID"
}

// wrongImplementation 演示使用 string key 的问题
func wrongImplementation() {
	fmt.Println("=== 错误示例：使用 string 作为 key ===")

	ctx := context.Background()

	// 第一层：库 A 设置 userID
	ctx = context.WithValue(ctx, wrongUserIDKey(), "user-001")
	fmt.Printf("库 A 设置后，userID: %v\n", ctx.Value(wrongUserIDKey()))

	// 第二层：库 B 也使用 "userID" 作为 key（覆盖了库 A 的值）
	ctx = context.WithValue(ctx, wrongUserIDKey(), "user-002")
	fmt.Printf("库 B 设置后，userID: %v\n", ctx.Value(wrongUserIDKey()))

	// 第三层：库 C 也使用 "userID"（再次覆盖）
	ctx = context.WithValue(ctx, wrongUserIDKey(), "user-003")
	fmt.Printf("库 C 设置后，userID: %v\n", ctx.Value(wrongUserIDKey()))

	// 问题：库 A 的原始值已经被覆盖，无法找回！
	fmt.Printf("\n严重问题：库 A 的原始值 'user-001' 已被覆盖！\n")
	fmt.Printf("当前值: %v（应该是被覆盖的值）\n\n", ctx.Value(wrongUserIDKey()))
}

// -----------------------------------------------------------------------------
// 正确示例：使用自定义 key 类型
// -----------------------------------------------------------------------------

// correctImplementation 演示使用自定义 key 类型的正确做法
func correctImplementation() {
	fmt.Println("=== 正确示例：使用自定义 key 类型 ===")

	ctx := context.Background()

	// 第一层：库 A 使用自定义 key 类型设置 userID
	ctx = context.WithValue(ctx, userIDKey, "user-001")
	fmt.Printf("库 A 设置后，userID: %v\n", ctx.Value(userIDKey))

	// 第二层：库 B 使用自己的 key（即使字符串也是 "userID"，但类型不同所以不冲突）
	ctx = context.WithValue(ctx, requestIDKey, "req-456")
	fmt.Printf("库 B 设置后，requestID: %v\n", ctx.Value(requestIDKey))
	fmt.Printf("库 A 的 userID 仍然安全: %v\n", ctx.Value(userIDKey))

	// 第三层：库 C 使用自己的 key
	ctx = context.WithValue(ctx, traceIDKey, "trace-789")
	fmt.Printf("库 C 设置后，traceID: %v\n", ctx.Value(traceIDKey))

	// 所有原始值都保持不变
	fmt.Printf("\n所有值都正确保留：\n")
	fmt.Printf("  userID: %v\n", ctx.Value(userIDKey))
	fmt.Printf("  requestID: %v\n", ctx.Value(requestIDKey))
	fmt.Printf("  traceID: %v\n\n", ctx.Value(traceIDKey))
}

// -----------------------------------------------------------------------------
// 类型安全的访问函数示例
// -----------------------------------------------------------------------------

// GetUserID 安全地获取 userID
func GetUserID(ctx context.Context) string {
	if val := ctx.Value(userIDKey); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetRequestID 安全地获取 requestID
func GetRequestID(ctx context.Context) string {
	if val := ctx.Value(requestIDKey); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetTraceID 安全地获取 traceID
func GetTraceID(ctx context.Context) string {
	if val := ctx.Value(traceIDKey); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// typeSafeAccess 演示类型安全的访问方式
func typeSafeAccess() {
	fmt.Println("=== 类型安全的访问函数示例 ===")

	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-123")
	ctx = context.WithValue(ctx, requestIDKey, "req-abc")
	ctx = context.WithValue(ctx, traceIDKey, "trace-xyz")

	// 使用类型安全的访问函数
	fmt.Printf("userID: %s\n", GetUserID(ctx))
	fmt.Printf("requestID: %s\n", GetRequestID(ctx))
	fmt.Printf("traceID: %s\n", GetTraceID(ctx))

	// 从没有设置值的 context 获取，返回空字符串
	emptyCtx := context.Background()
	fmt.Printf("空 context 的 userID: '%s'\n\n", GetUserID(emptyCtx))
}

// -----------------------------------------------------------------------------
// Context 链路查找示例
// -----------------------------------------------------------------------------

// contextChainLookup 演示 context value 的查找机制
func contextChainLookup() {
	fmt.Println("=== Context 链路查找示例 ===")

	ctx := context.Background()

	// 初始 context 没有值
	fmt.Printf("初始 context，userID: %v\n", ctx.Value(userIDKey))

	// 第一层：设置 userID
	ctx = context.WithValue(ctx, userIDKey, "user-001")
	fmt.Printf("第一层 context，userID: %v\n", ctx.Value(userIDKey))

	// 第二层：基于第一层创建子 context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 子 context 继承了父 context 的值
	fmt.Printf("子 context（继承），userID: %v\n", ctx.Value(userIDKey))

	// 第三层：在子 context 中覆盖 userID
	ctx = context.WithValue(ctx, userIDKey, "user-002")
	fmt.Printf("子 context（覆盖后），userID: %v\n", ctx.Value(userIDKey))

	// 父 context 的值不受影响
	fmt.Printf("父 context，userID: %v（未受影响）\n\n", GetUserID(context.Background()))
}

// -----------------------------------------------------------------------------
// 模拟真实场景：HTTP 请求追踪
// -----------------------------------------------------------------------------

// simulateHTTPRequest 模拟在 HTTP 请求中使用 context 传递追踪信息
func simulateHTTPRequest() {
	fmt.Println("=== 模拟 HTTP 请求追踪场景 ===")

	// 模拟从请求入口设置追踪信息
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-001")
	ctx = context.WithValue(ctx, requestIDKey, "req-12345")
	ctx = context.WithValue(ctx, traceIDKey, "trace-abc")

	// 模拟 handler 处理
	handleRequest(ctx)

	fmt.Println()
}

// handleRequest 模拟请求处理器
func handleRequest(ctx context.Context) {
	fmt.Printf("处理请求 - userID: %s, requestID: %s, traceID: %s\n",
		GetUserID(ctx), GetRequestID(ctx), GetTraceID(ctx))

	// 模拟调用下游服务
	callDownstreamService(ctx)
}

// callDownstreamService 模拟调用下游服务
func callDownstreamService(ctx context.Context) {
	// 下游服务也能正确获取追踪信息
	fmt.Printf("下游服务 - userID: %s, requestID: %s, traceID: %s\n",
		GetUserID(ctx), GetRequestID(ctx), GetTraceID(ctx))
}

// -----------------------------------------------------------------------------
// 主函数：运行所有示例
// -----------------------------------------------------------------------------

func main() {
	fmt.Println("================================================================")
	fmt.Println("              Context Values 示例 - 演示 key 冲突问题            ")
	fmt.Println("================================================================\n")

	// 示例 1：错误示例
	wrongImplementation()

	// 示例 2：正确示例
	correctImplementation()

	// 示例 3：类型安全的访问
	typeSafeAccess()

	// 示例 4：Context 链路查找
	contextChainLookup()

	// 示例 5：模拟真实场景
	simulateHTTPRequest()

	fmt.Println("=== 示例运行完成 ===")
}
