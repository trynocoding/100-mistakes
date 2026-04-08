//go:build ignore

package main

import (
	"fmt"
	"math"
	"runtime"
	"strings"

	"go.uber.org/automaxprocs/maxprocs"
)

// 本演示揭示 automaxprocs 与 Go 原生容器感知 GOMAXPROCS 的核心差异
//
// 关键差异：
// 1. 取整函数：automaxprocs 用 Floor，Go 原生用 Ceil
// 2. 最小值：automaxprocs 默认 1，Go 原生默认 2

func main() {
	fmt.Println("=== automaxprocs vs Go 原生：核心差异分析 ===")
	fmt.Println()

	// 1. 核心差异：取整函数
	fmt.Println("--- 差异 1：取整函数 ---")
	fmt.Println("automaxprocs DefaultRoundFunc: math.Floor (向下取整)")
	fmt.Println("Go 原生 adjustCgroupGOMAXPROCS: ceil (向上取整)")
	fmt.Println()

	quota := 0.5 // 500m CPU
	fmt.Printf("假设 quota = %.1f (500m CPU):\n", quota)
	fmt.Printf("  automaxprocs: Floor(%.1f) = %.1f → int = %d\n", quota, math.Floor(quota), int(math.Floor(quota)))
	fmt.Printf("  Go 原生:     Ceil(%.1f)  = %.1f → max(%.1f, 2) = 2\n", quota, math.Ceil(quota), math.Ceil(quota))
	fmt.Println()

	// 2. 最小值差异
	fmt.Println("--- 差异 2：最小值 ---")
	fmt.Println("automaxprocs: minGOMAXPROCS = 1 (可通过 maxprocs.Min() 调整)")
	fmt.Println("Go 原生:     min = 2 (硬编码)")
	fmt.Println()

	// 3. 计算结果对比
	fmt.Println("--- 差异 3：实际计算结果 ---")
	headers := []string{"CPU Limit", "automaxprocs (Floor→1)", "Go 原生 (Ceil→2)"}
	printRow(headers)
	printDivider()

	testCases := []float64{0.25, 0.5, 1.0, 1.5, 2.0, 3.0}
	for _, q := range testCases {
		autoResult := int(math.Floor(q))
		if autoResult < 1 {
			autoResult = 1
		}

		goResult := math.Ceil(q)
		if goResult < 2 {
			goResult = 2
		}

		autoStr := fmt.Sprintf("%d", autoResult)
		goStr := fmt.Sprintf("%.0f", goResult)
		quotaStr := fmt.Sprintf("%.2f (%dm)", q, int(q*1000))
		printRow([]string{quotaStr, autoStr, goStr})
	}
	fmt.Println()

	// 4. 环境变量行为差异
	fmt.Println("--- 差异 4：环境变量处理 ---")
	fmt.Println("automaxprocs: 检查 GOMAXPROCS 环境变量，如果已设置则不覆盖")
	fmt.Println("Go 原生:     不检查环境变量，直接设置")
	fmt.Println()
	fmt.Println("automaxprocs 源码 (maxprocs.go:109):")
	fmt.Println(`  if max, exists := os.LookupEnv(_maxProcsKey); exists {`)
	fmt.Println(`      cfg.log("maxprocs: Honoring GOMAXPROCS=%q as set in environment", max)`)
	fmt.Println(`      return undoNoop, nil`)
	fmt.Println(`  }`)
	fmt.Println()

	// 5. Undo 功能（automaxprocs 独有）
	fmt.Println("--- automaxprocs 独有特性：Undo ---")
	fmt.Println("automaxprocs.Set() 返回一个 undo() 函数，可以恢复之前的 GOMAXPROCS 值")
	fmt.Println()

	// 演示 Undo
	current := runtime.GOMAXPROCS(0)
	undo, _ := maxprocs.Set()
	newVal := runtime.GOMAXPROCS(0)
	fmt.Printf("设置前: GOMAXPROCS = %d\n", current)
	fmt.Printf("设置后: GOMAXPROCS = %d\n", newVal)
	undo()
	restored := runtime.GOMAXPROCS(0)
	fmt.Printf("Undo后: GOMAXPROCS = %d\n", restored)
	fmt.Println()

	// 6. Min 选项（automaxprocs 独有）
	fmt.Println("--- automaxprocs 独有特性：Min 选项 ---")
	fmt.Println("maxprocs.Min(n) 可以设置最小值覆盖默认的 1")
	fmt.Println("Go 原生不支持此选项")
	fmt.Println()

	// 演示 Min
	current = runtime.GOMAXPROCS(0)
	undo, _ = maxprocs.Set(maxprocs.Min(4))
	newVal = runtime.GOMAXPROCS(0)
	fmt.Printf("设置前: GOMAXPROCS = %d\n", current)
	fmt.Printf("使用 maxprocs.Min(4) 后: GOMAXPROCS = %d\n", newVal)
	undo()
	fmt.Println()

	// 7. 总结
	printSummary()
}

func printSummary() {
	fmt.Println("=== 总结：automaxprocs vs Go 原生 ===")
	fmt.Println()

	headers := []string{"特性", "automaxprocs", "Go 原生 (1.25+)"}
	printRow(headers)
	printDivider()

	rows := [][]string{
		{"取整函数", "math.Floor (向下)", "math.Ceil (向上)"},
		{"默认最小值", "1", "2"},
		{"500m → GOMAXPROCS", "1 (Floor→0→Min→1)", "2 (Ceil→1→Max→2)"},
		{"Min 选项", "✓ 支持", "✗ 不支持"},
		{"Undo 恢复", "✓ 支持", "✗ 不支持"},
		{"Logger 日志", "✓ 支持", "✗ 不支持"},
		{"环境变量检查", "✓ 已有则跳过", "✗ 不检查"},
		{"GODEBUG 禁用", "N/A", "✓ containermaxprocs=0"},
		{"Go < 1.25 支持", "✓", "✗"},
	}

	for _, row := range rows {
		printRow(row)
	}

	fmt.Println()
	fmt.Println("重要结论：")
	fmt.Println("- 相同 CPU quota 下，automaxprocs 可能得到比 Go 原生更小的值")
	fmt.Println("- Go 原生保证最小值为 2，automaxprocs 默认最小值为 1")
	fmt.Println("- Go 1.25+ 环境如需日志和 Undo，建议使用 automaxprocs")
}

func printRow(cols []string) {
	widths := []int{25, 28, 25}
	for i, col := range cols {
		fmt.Printf("%-*s", widths[i], col)
	}
	fmt.Println()
}

func printDivider() {
	fmt.Println(strings.Repeat("-", 78))
}
