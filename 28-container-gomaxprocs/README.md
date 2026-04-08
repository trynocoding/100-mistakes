# 容器中的 GOMAXPROCS

## 问题背景

在 **Go 1.25 之前**，在容器化环境中运行 Go 应用时，`runtime.GOMAXPROCS()` 函数会返回宿主机的 CPU 核心数，而不是容器被限制的 CPU 核心数。

**从 Go 1.25 起**，Go runtime 已经**内置**了容器感知的 GOMAXPROCS 自动调整机制，不再需要依赖外部库。

---

## Go 1.25+ 的内置行为

### 核心规则

Go 1.25+ 的 runtime 在启动时会自动检测容器 cgroup CPU 限制，`GOMAXPROCS` 默认值按以下公式计算：

```
GOMAXPROCS = min(sched_getaffinity可见CPU数, max(ceil(quota/period), 2))
```

### 公式解释

1. **`sched_getaffinity` 可见的逻辑 CPU 数**：宿主机视角的可用 CPU 数
2. **cgroup CPU 吞吐限制**：`quota / period`
   - 小于 2 的值会提升到 2（最小保底值）
   - 带小数的值会向上取整
3. **取两者中的较小值**

### 计算示例

假设容器 CPU limit 为 `500m`（0.5 CPU）：

```
quota / period ≈ 0.5
ceil(0.5) = 1
max(1, 2) = 2
min(8, 2) = 2
```

即使容器限制是 0.5 CPU，`GOMAXPROCS` 也会是 2（因为最小保底值是 2）。

### cgroup 信息来源

- **cgroup v2**：读取 `cpu.max`
- **cgroup v1**：读取 `cpu.cfs_quota_us` 和 `cpu.cfs_period_us`

---

## 实际验证

```go
// 在容器中运行（容器限制 500m）
fmt.Println("NumCPU:     =", runtime.NumCPU())      // 输出: 8（宿主机逻辑CPU）
fmt.Println("GOMAXPROCS =", runtime.GOMAXPROCS(0))   // 输出: 2（cgroup限制）
```

这正是 Go 1.25+ runtime 的预期行为。

---

## 已知局限

### 1. 只读取叶子 cgroup

Go runtime 只读取**当前进程所在叶子 cgroup** 的 CPU limit，**不会检查父 cgroup**。

```go
// We only read the limit from the leaf cgroup that actually contains this
// process. But a parent cgroup may have a tighter limit. That tighter limit
// would be our effective limit.
```

这意味着如果父 cgroup 有更严格的限制，它不会被应用。不过容器运行时通常会把父 cgroup 隐藏。

### 2. 进程迁移到其他 cgroup 不生效

cgroup 检测只在**启动时执行一次**：

```go
// If the process is migrated to another cgroup while it is running it will
// not notice, as we only check which cgroup we are in once at startup.
```

如果进程在运行期间被迁移到其他 cgroup，Go runtime 不会感知到新的 CPU limit。

### 3. 可通过 GODEBUG 禁用

```bash
GODEBUG=containermaxprocs=0 ./your_program
```

```go
if debug.containermaxprocs > 0 {
    // Normal operation - 启用 cgroup 感知
    cgroupCPU = c
    cgroupOK = true
    return
}
// cgroup-aware GOMAXPROCS is disabled.
```

禁用后，`GOMAXPROCS` 将只基于 `sched_getaffinity` 返回的 CPU 数，不再受 cgroup limit 限制。

---

## 何时仍需要 automaxprocs

`go.uber.org/automaxprocs` 在以下场景仍有价值：

1. **Go 版本 < 1.25**：需要手动设置正确的 GOMAXPROCS
2. **需要显式控制**：不想依赖 runtime 默认行为
3. **特殊 cgroup 配置**：非标准 cgroup 路径或层级
4. **需要日志记录**：查看 GOMAXPROCS 是如何被计算的

### 安装

```bash
go get go.uber.org/automaxprocs
```

### 使用方法

```go
import "go.uber.org/automaxprocs"

func main() {
    // 包会自动在 init() 中根据 cgroup 配置设置 GOMAXPROCS
    fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))
}
```

### 手动控制

```go
import (
    "go.uber.org/automaxprocs/maxprocs"
    "log"
)

func main() {
    undo, err := maxprocs.Set(maxprocs.Logger(log.Printf))
    if err != nil {
        log.Printf("maxprocs warning: %v", err)
    }
    defer undo()

    fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))
}
```

---

## automaxprocs 与 Go 原生对比

参见 `automaxprocs_demo.go`，演示两者的核心差异：

### 功能对比表

| 特性 | automaxprocs | Go 原生 (1.25+) |
|-----|-------------|-----------------|
| 容器 CPU limit 感知 | ✓ | ✓ |
| 最小保底值 2 | ✓ | ✓ |
| 小数向上取整 | ✓ | ✓ |
| 启动时自动执行 | ✓ (init) | ✓ (runtime) |
| **日志输出** | ✓ Logger | ✗ |
| **Undo 恢复** | ✓ | ✗ |
| **GODEBUG 禁用** | N/A | ✓ |
| **Go < 1.25 支持** | ✓ | ✗ |

### 核心结论

- **Go 1.25+ 环境**：automaxprocs 的主要价值是**日志可见性** + **Undo 恢复能力**，核心计算逻辑与 Go 原生一致
- **Go < 1.25 环境**：必须使用 automaxprocs 才能获得容器感知能力

---

## 常见误区

### 误区 1：NumCPU 和 GOMAXPROCS 是同一个值

**错误**。在 Go 1.25+ 容器环境中：
- `NumCPU()` = 宿主机可见的逻辑 CPU 数（通常是 8）
- `GOMAXPROCS(0)` = cgroup 限制调整后的值（通常是 2）

### 误区 2：GOMAXPROCS=1 表示单核

**错误**。GOMAXPROCS=1 只表示运行时同时最多调度 1 个 OS 线程来执行 goroutine，但 goroutine 总数不受限制。

### 误区 3：500m 限制会得到 GOMAXPROCS=1

**错误**。由于最小保底值是 2，所以 `500m`、`250m` 等小于 2 的限制都会得到 `GOMAXPROCS=2`。

---

## 相关阅读

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Container-aware GOMAXPROCS](https://go.dev/blog/container-aware-gomaxprocs)
- [Go runtime cgroup_linux.go](https://go.dev/src/runtime/cgroup_linux.go)
- [uber-go/automaxprocs 官方文档](https://pkg.go.dev/go.uber.org/automaxprocs)
