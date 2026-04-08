//go:build ignore

package main

import (
	"fmt"
	"runtime"
)

// 注意：Go 1.25+ 的 runtime 已经内置了容器感知的 GOMAXPROCS。
// 无需外部库，runtime 启动时就会自动根据 cgroup CPU limit 调整 GOMAXPROCS。
//
// 本示例演示 Go 1.25+ 的内置行为。

func main() {
	fmt.Println("=== Go 1.25+ 容器 GOMAXPROCS 行为 ===")
	fmt.Println()

	// NumCPU() 返回当前进程通过 sched_getaffinity 可见的逻辑 CPU 数
	// 这是宿主机视角的值，不受 cgroup 限制影响
	fmt.Printf("runtime.NumCPU(): %d\n", runtime.NumCPU())
	fmt.Printf("  (这是宿主机可见的逻辑 CPU 数)\n")
	fmt.Println()

	// GOMAXPROCS(0) 返回当前 GOMAXPROCS 设置
	// 在 Go 1.25+ 容器环境中，这会自动被 cgroup CPU limit 限制
	//
	// 计算公式：min(NumCPU, max(ceil(quota/period), 2))
	//
	// 例如容器限制为 500m (0.5 CPU)：
	//   quota/period ≈ 0.5 → ceil(0.5) = 1 → max(1, 2) = 2
	//   min(8, 2) = 2
	fmt.Printf("runtime.GOMAXPROCS(0): %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("  (这是 runtime 根据 cgroup limit 自动调整后的值)\n")
	fmt.Println()

	// 手动设置 GOMAXPROCS 会覆盖默认值
	// 设置后可以通过 GOMAXPROCS(0) 查看当前值
	original := runtime.GOMAXPROCS(4)
	fmt.Printf("手动设置 GOMAXPROCS=4，原值: %d\n", original)
	fmt.Printf("设置后 runtime.GOMAXPROCS(0): %d\n", runtime.GOMAXPROCS(0))

	// 恢复为默认值（让 runtime 重新计算）
	runtime.GOMAXPROCS(original)
	fmt.Printf("恢复后 runtime.GOMAXPROCS(0): %d\n", runtime.GOMAXPROCS(0))
}
