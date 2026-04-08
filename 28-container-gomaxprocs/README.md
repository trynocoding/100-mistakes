# 容器中的 GOMAXPROCS

## 问题背景

在容器化环境中运行 Go 应用时，`runtime.GOMAXPROCS()` 函数会返回**宿主机**的 CPU 核心数，而不是容器被限制的 CPU 核心数。这会导致以下问题：

1. **资源超限**：Go 运行时创建的 goroutine 数量基于宿主机核数，可能超出容器 cgroup 限制
2. **性能下降**：频繁的 CPU 竞争和上下文切换
3. **资源浪费**：无法准确评估容器需要的资源配额
4. **调度不准确**：`runtime.NumGoroutine()` 和 `GOMAXPROCS()` 不一致

### 问题示例

假设宿主机有 32 核，容器被限制为 2 核：

```go
// 容器中运行
fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))        // 输出: 32（错误！）
fmt.Println("NumCPU:", runtime.NumCPU())                  // 输出: 32（错误！）
fmt.Println("容器 cgroup 限制: 2 核")
```

实际上 Go 运行时会尝试在这 32 个 CPU 上调度 goroutine，但容器的 cgroup 限制只有 2 核，导致资源不匹配。

### 根本原因

`runtime.GOMAXPROCS()` 和 `runtime.NumCPU()` 底层调用的是系统的 `sched_getaffinity()`（Linux）或类似接口，这些接口返回的是宿主机视角的 CPU 核心数，而不是容器 cgroup 视角的限制。

---

## 解决方案

使用 `go.uber.org/automaxprocs/maxprocs` 库，它会自动读取容器 cgroup 配置并设置正确的 GOMAXPROCS 值。

### 安装

```bash
go get go.uber.org/automaxprocs
```

### 使用方法

#### 方式一：自动设置（推荐）

只需导入包，即可在 `init()` 函数中自动完成设置：

```go
import "go.uber.org/automaxprocs"

func main() {
    // 包会自动在 init() 中根据 cgroup 配置设置 GOMAXPROCS
    fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))
}
```

#### 方式二：手动控制（带日志）

使用 `maxprocs.Set()` 函数进行手动控制，并可选地设置日志回调：

```go
import (
    "go.uber.org/automaxprocs/maxprocs"
    "log"
)

func main() {
    // Set 会自动读取 cgroup 配置并设置 GOMAXPROCS
    undo, err := maxprocs.Set(maxprocs.Logger(log.Printf))
    if err != nil {
        // 在非容器环境或无法获取 cgroup 信息时，err 可能不为 nil
        log.Printf("maxprocs warning: %v", err)
    }
    defer undo() // 可以通过 undo() 恢复之前的设置

    fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))
}
```

---

## 注意事项

1. **调用时机**：使用 `maxprocs.Set()` 时，必须在 Go 运行时初始化后、任何需要并行度的操作之前调用
2. **依赖库**：如果使用了支持 automaxprocs 的库（如 `uber-go/goleak`），可能已自动集成
3. **Kubernetes**：在 K8s 环境中，cgroup 限制通常与 CPU 请求/限制一致，automaxprocs 可以正确识别
4. **本地开发**：在非容器环境中，automaxprocs 不会做任何修改，保持原有行为

---

## 验证方法

运行示例代码，观察输出：
- 在容器中：`GOMAXPROCS` 应等于容器 CPU 限制
- 本地环境：应等于机器实际核数

---

## 相关阅读

- [uber-go/automaxprocs 官方文档](https://pkg.go.dev/go.uber.org/automaxprocs)
- [Go runtime NumCPU 文档](https://pkg.go.dev/runtime#NumCPU)
- [Linux cgroups CPU 子系统](https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v2.html)
