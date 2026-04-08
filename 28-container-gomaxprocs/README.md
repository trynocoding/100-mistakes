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

### 核心差异：取整函数和最小值

| 特性 | automaxprocs | Go 原生 (1.25+) |
|-----|-------------|-----------------|
| **取整函数** | `math.Floor`（向下取整） | `math.Ceil`（向上取整） |
| **默认最小值** | 1 | 2 |
| **Min 选项** | ✓ 支持 | ✗ 不支持 |
| **Undo 恢复** | ✓ 支持 | ✗ 不支持 |
| **日志输出** | ✓ Logger | ✗ 不支持 |
| **环境变量检查** | ✓ 已有则跳过 | ✗ 不检查 |
| **GODEBUG 禁用** | N/A | ✓ |
| **Go < 1.25 支持** | ✓ | ✗ |

### 计算结果对比

假设 `NumCPU = 8`：

| CPU Limit | automaxprocs 计算 | Go 原生计算 |
|-----------|-------------------|-------------|
| 250m (0.25) | `Floor(0.25)=0` → `Min(1)` = **1** | `Ceil(0.25)=1` → `Max(2)` = **2** |
| 500m (0.5) | `Floor(0.5)=0` → `Min(1)` = **1** | `Ceil(0.5)=1` → `Max(2)` = **2** |
| 1000m (1.0) | `Floor(1.0)=1` = **1** | `Ceil(1.0)=1` → `Max(2)` = **2** |
| 1500m (1.5) | `Floor(1.5)=1` = **1** | `Ceil(1.5)=2` → `Max(2)` = **2** |
| 2000m (2.0) | `Floor(2.0)=2` = **2** | `Ceil(2.0)=2` = **2** |
| 3000m (3.0) | `Floor(3.0)=3` = **3** | `Ceil(3.0)=3` = **3** |

**关键发现**：对于小于 2 CPU 的限制，Go 原生**总是返回 2**，而 automaxprocs 可能返回 1。

### automaxprocs 独有特性

#### Min 选项
```go
// 设置最小值为 4，即使 quota 只有 0.5
undo, _ := maxprocs.Set(maxprocs.Min(4))
```

#### Undo 恢复
```go
undo, _ := maxprocs.Set()
defer undo() // 恢复到设置前的值
```

#### 环境变量兼容
```go
// 如果设置了 GOMAXPROCS 环境变量，automaxprocs 不会覆盖
// export GOMAXPROCS=4
undo, _ := maxprocs.Set() // 不会修改，保持 4
```

### 核心结论

1. **Go 1.25+ 环境**：两者计算逻辑**并不完全一致**，Go 原生更保守（最小值 2）
2. **需要日志/Undo**：使用 automaxprocs
3. **需要保守策略**：Go 原生保证最小 2，适合资源受限场景
4. **Go < 1.25**：必须使用 automaxprocs

---

## 常见误区

### 误区 1：NumCPU 和 GOMAXPROCS 是同一个值

**错误**。在 Go 1.25+ 容器环境中：
- `NumCPU()` = 宿主机可见的逻辑 CPU 数（通常是 8）
- `GOMAXPROCS(0)` = cgroup 限制调整后的值（通常是 2）

### 误区 2：GOMAXPROCS=1 表示单核

**错误**。GOMAXPROCS=1 只表示运行时同时最多调度 1 个 OS 线程来执行 goroutine，但 goroutine 总数不受限制。

### 误区 3：automaxprocs 和 Go 原生计算结果相同

**错误**。两者使用不同的取整函数和最小值：

- **automaxprocs**：`Floor` + 最小值 1 → 500m 得到 **GOMAXPROCS=1**
- **Go 原生**：`Ceil` + 最小值 2 → 500m 得到 **GOMAXPROCS=2**

这是有意设计的差异：Go 原生更保守，确保即使极小 CPU 限制也能有至少 2 个线程。

---

## 相关阅读

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Container-aware GOMAXPROCS](https://go.dev/blog/container-aware-gomaxprocs)
- [Go runtime cgroup_linux.go](https://go.dev/src/runtime/cgroup_linux.go)
- [uber-go/automaxprocs 官方文档](https://pkg.go.dev/go.uber.org/automaxprocs)
