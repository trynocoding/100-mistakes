package main

import (
	"errors"
	"fmt"
)

// =============================================================================
// Break 作用域问题示例代码
// Go 1.21+ 可运行
// =============================================================================

// -----------------------------------------------------------------------------
// 示例 1：break 只跳出 switch，不跳出 for 循环（错误示例）
// -----------------------------------------------------------------------------

// findValueWrong 错误演示：break 只跳出 switch
func findValueWrong(matrix [][]int, target int) bool {
	fmt.Printf("错误示例：查找目标值 %d\n", target)
	for rowIdx, row := range matrix {
		for colIdx, col := range row {
			switch col {
			case target:
				// 错误：这里的 break 只会跳出 switch，不会跳出外层 for 循环
				// 程序会继续执行外层循环的下一次迭代
				break
			}
			fmt.Printf("  位置 [%d][%d]: %d (不是目标)\n", rowIdx, colIdx, col)
		}
	}
	fmt.Println("  循环结束，未找到目标")
	return false
}

// findValueRight 正确演示：使用 return 跳出所有循环
func findValueRight(matrix [][]int, target int) bool {
	fmt.Printf("\n正确示例（使用 return）：查找目标值 %d\n", target)
	for rowIdx, row := range matrix {
		for colIdx, col := range row {
			switch col {
			case target:
				fmt.Printf("  找到目标！位置 [%d][%d]: %d\n", rowIdx, colIdx, col)
				return true
			}
			fmt.Printf("  位置 [%d][%d]: %d (不是目标)\n", rowIdx, colIdx, col)
		}
	}
	fmt.Println("  未找到目标")
	return false
}

// -----------------------------------------------------------------------------
// 示例 2：使用 label 正确跳出外层循环
// -----------------------------------------------------------------------------

// findValueWithLabel 使用 label 跳出外层循环
func findValueWithLabel(matrix [][]int, target int) bool {
	fmt.Printf("\n正确示例（使用 label）：查找目标值 %d\n", target)
loop:
	for rowIdx, row := range matrix {
		for colIdx, col := range row {
			switch col {
			case target:
				fmt.Printf("  找到目标！位置 [%d][%d]: %d\n", rowIdx, colIdx, col)
				break loop // 使用 label 跳出外层循环
			}
			fmt.Printf("  位置 [%d][%d]: %d (不是目标)\n", rowIdx, colIdx, col)
		}
	}
	fmt.Println("  循环结束（通过 label 跳出）")
	return false
}

// -----------------------------------------------------------------------------
// 示例 3：break 无法跳出外层 for 循环的对比
// -----------------------------------------------------------------------------

// badBreakInNestedLoop 错误示例：普通 break 无法跳出外层循环
func badBreakInNestedLoop() int {
	fmt.Println("\n错误示例：普通 break 无法跳出外层循环")
	result := 0

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			switch {
			case i*j > 4:
				// 错误：这里 break 只会跳出 switch
				// 外层 for 循环继续执行
				break
			}
			fmt.Printf("  i=%d, j=%d, i*j=%d\n", i, j, i*j)
		}
	}
	return result
}

// goodBreakInNestedLoop 正确示例：使用 label 跳出外层循环
func goodBreakInNestedLoop() int {
	fmt.Println("\n正确示例：使用 break label 跳出外层循环")
	result := 0

loop:
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			switch {
			case i*j > 4:
				result = i*j
				// 正确：使用 break loop 跳出外层 for 循环
				break loop
			}
			fmt.Printf("  i=%d, j=%d, i*j=%d\n", i, j, i*j)
		}
	}
	fmt.Printf("  最终结果: %d\n", result)
	return result
}

// -----------------------------------------------------------------------------
// 示例 4：嵌套 switch 中的 label 使用
// -----------------------------------------------------------------------------

// nestedSwitchDemo 嵌套 switch 中的 label 使用
func nestedSwitchDemo(n int) {
	fmt.Printf("\n嵌套 switch 示例（n=%d）\n", n)

loop:
	for i := 0; i < n; i++ {
		fmt.Printf("  外层循环: i=%d\n", i)
		switch i {
		case 2:
			for j := 0; j < n; j++ {
				fmt.Printf("    内层循环: j=%d\n", j)
				switch j {
				case 1:
					// 当 i=2, j=1 时，跳出外层 for 循环
					fmt.Println("    条件满足，跳出外层循环")
					break loop
				}
			}
		}
	}
	fmt.Println("  嵌套 switch 示例结束")
}

// -----------------------------------------------------------------------------
// 示例 5：模拟实际场景 - 验证用户输入
// -----------------------------------------------------------------------------

// validateUserInput 模拟验证用户输入的场景
func validateUserInput(input string) error {
	fmt.Printf("\n验证用户输入: %q\n", input)

	for _, char := range input {
		switch {
		case char >= 'a' && char <= 'z':
			fmt.Printf("  小写字母: %c\n", char)
		case char >= 'A' && char <= 'Z':
			fmt.Printf("  大写字母: %c\n", char)
		case char >= '0' && char <= '9':
			fmt.Printf("  数字: %c\n", char)
		default:
			// 这里使用 return 直接返回错误
			// 因为 break 只能跳出 switch，不能跳出 for
			return fmt.Errorf("遇到不支持的字符: %c", char)
		}
	}
	return nil
}

// validateUserInputWithLabel 使用 label 的方式（功能相同但结构不同）
func validateUserInputWithLabel(input string) error {
	fmt.Printf("\n使用 label 验证用户输入: %q\n", input)

loop:
	for _, char := range input {
		switch {
		case char >= 'a' && char <= 'z':
			fmt.Printf("  小写字母: %c\n", char)
		case char >= 'A' && char <= 'Z':
			fmt.Printf("  大写字母: %c\n", char)
		case char >= '0' && char <= '9':
			fmt.Printf("  数字: %c\n", char)
		default:
			fmt.Printf("  遇到不支持的字符: %c，跳出循环\n", char)
			break loop
		}
	}
	return errors.New("输入包含不支持的字符")
}

// -----------------------------------------------------------------------------
// 主函数：运行所有示例
// -----------------------------------------------------------------------------

func main() {
	fmt.Println("=== Break 作用域问题示例 ===\n")

	// 准备测试数据
	matrix := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	// 示例 1：break 只跳出 switch 的问题
	fmt.Println("--- 示例 1: break 只跳出 switch ---")
	findValueWrong(matrix, 5)
	findValueRight(matrix, 5)
	findValueWithLabel(matrix, 5)

	// 示例 2：break 无法跳出外层 for 循环的对比
	fmt.Println("\n--- 示例 2: break 与 break label 对比 ---")
	badBreakInNestedLoop()
	goodBreakInNestedLoop()

	// 示例 3：嵌套 switch 中的 label 使用
	fmt.Println("\n--- 示例 3: 嵌套 switch ---")
	nestedSwitchDemo(3)

	// 示例 4：实际场景 - 验证用户输入
	fmt.Println("\n--- 示例 4: 验证用户输入 ---")
	if err := validateUserInput("abc123"); err != nil {
		fmt.Printf("  错误: %v\n", err)
	}

	fmt.Println("\n=== 示例运行完成 ===")
}