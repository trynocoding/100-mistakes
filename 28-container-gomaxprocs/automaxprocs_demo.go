//go:build ignore

package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"go.uber.org/automaxprocs/maxprocs"
)

// 本演示展示 go.uber.org/automaxprocs 与 Go 原生容器感知 GOMAXPROCS 的区别
//
// Go 1.25+ runtime 已内置容器感知 GOMAXPROCS，
// automaxprocs 的价值在于：日志记录、Undo 恢复、GODEBUG 控制

func main() {
	fmt.Println("=== automaxprocs vs Go 原生容器感知 GOMAXPROCS ===")
	fmt.Println()

	// 1. Go 原生行为（Go 1.25+）
	fmt.Println("--- 1. Go 1.25+ 原生容器感知 ---")
	fmt.Printf("runtime.NumCPU(): %d\n", runtime.NumCPU())
	fmt.Printf("runtime.GOMAXPROCS(0): %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()
	fmt.Println("说明：Go 1.25+ runtime 启动时自动读取 cgroup CPU limit")
	fmt.Println("      公式：min(NumCPU, max(ceil(quota/period), 2))")
	fmt.Println()

	// 2. automaxprocs.Set 的核心差异
	fmt.Println("--- 2. automaxprocs.Set ---")

	// maxprocs.Set 返回一个 undo 函数，可以恢复之前的设置
	undo, err := maxprocs.Set(
		// Logger 选项：打印 automaxprocs 的设置过程
		// Go 原生不支持日志输出
		maxprocs.Logger(func(format string, args ...any) {
			log.Printf("[automaxprocs] "+format, args...)
		}),
	)
	if err != nil {
		// 在非容器环境或无法获取 cgroup 信息时，err 可能不为 nil
		log.Printf("[automaxprocs] warning: %v", err)
	}

	fmt.Printf("设置后 runtime.GOMAXPROCS(0): %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	// 3. Undo 功能：恢复之前的设置
	fmt.Println("--- 3. Undo 功能（automaxprocs 独有）---")
	fmt.Println("调用 undo() 恢复之前的 GOMAXPROCS 设置...")

	undo() // 恢复

	fmt.Printf("恢复后 runtime.GOMAXPROCS(0): %d\n", runtime.GOMAXPROCS(0))
	fmt.Println("Go 原生不支持 Undo，只能手动重新设置")
	fmt.Println()

	// 4. 手动重新设置
	fmt.Println("--- 4. 手动设置 GOMAXPROCS ---")
	original := runtime.GOMAXPROCS(0)
	runtime.GOMAXPROCS(4)
	fmt.Printf("手动设置 GOMAXPROCS=4，当前值: %d\n", runtime.GOMAXPROCS(0))

	// 使用 automaxprocs 恢复到"自动计算"的值
	undoAgain, _ := maxprocs.Set()
	fmt.Printf("使用 maxprocs.Set() 重新自动计算: %d\n", runtime.GOMAXPROCS(0))
	undoAgain()

	// 恢复原始值
	runtime.GOMAXPROCS(original)
	fmt.Println()

	// 5. GODEBUG 控制（Go 原生）
	fmt.Println("--- 5. GODEBUG=containermaxprocs=0（仅 Go 原生） ---")
	fmt.Println("这是 Go 原生提供的环境变量禁用方式")
	fmt.Println("automaxprocs 不提供此功能（因为它本身就是替代方案）")
	fmt.Println()
	fmt.Println("使用方式：GODEBUG=containermaxprocs=0 ./your_program")
	fmt.Println()

	// 6. 总结对比
	printSummary()
}

func printSummary() {
	fmt.Println("=== 总结：automaxprocs vs Go 原生 ===")
	fmt.Println()

	headers := []string{"特性", "automaxprocs", "Go 原生 (1.25+)"}
	printRow(headers)
	printDivider()

	rows := [][]string{
		{"容器 CPU limit 感知", "✓", "✓"},
		{"最小保底值 2", "✓", "✓"},
		{"小数向上取整", "✓", "✓"},
		{"启动时自动执行", "✓ (init)", "✓ (runtime)"},
		{"日志输出", "✓ Logger", "✗"},
		{"Undo 恢复", "✓", "✗"},
		{"手动 Set() 控制", "✓", "✗ (需手动调)"},
		{"GODEBUG 禁用", "N/A", "✓"},
		{"Go < 1.25 支持", "✓", "✗"},
	}

	for _, row := range rows {
		printRow(row)
	}

	fmt.Println()
	fmt.Println("结论：Go 1.25+ 环境下，automaxprocs 的主要价值是")
	fmt.Println("      日志可见性 + Undo 恢复能力。核心计算逻辑一致。")
}

func printRow(cols []string) {
	// 计算每列宽度
	widths := []int{25, 18, 18}
	for i, col := range cols {
		fmt.Printf("%-*s", widths[i], col)
	}
	fmt.Println()
}

func printDivider() {
	fmt.Println(strings.Repeat("-", 61))
}
