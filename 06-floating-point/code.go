// +build ignore

package main

import (
	"fmt"
	"log"
	"math"

	"github.com/shopspring/decimal"
)

// 本文件演示 Go 语言中的浮点数精度问题
// 运行方式：go run code.go
// 需要先安装 decimal 库：go get github.com/shopspring/decimal

func main() {
	fmt.Println("=== 浮点数精度问题演示 ===\n")

	// 1. 基本精度丢失示例
	basicPrecisionLoss()

	// 2. 先乘后加 vs 先加后乘
	orderOfOperations()

	// 3. 大数吃小数问题
	largeNumberEatsSmall()

	// 4. 使用 decimal 库解决精度问题
	decimalExample()
}

// basicPrecisionLoss 演示 0.1 + 0.2 不等于 0.3
func basicPrecisionLoss() {
	fmt.Println("--- 基本精度丢失示例 ---")

	a := 0.1
	b := 0.2
	sum := a + b

	fmt.Printf("0.1 + 0.2 = %.20f\n", sum)
	fmt.Printf("0.3       = %.20f\n", 0.3)
	fmt.Printf("差值:     %.20e\n", math.Abs(sum-0.3))

	// 错误比较方式
	if sum == 0.3 {
		fmt.Println("sum == 0.3: true (错误，实际上不相等)")
	} else {
		fmt.Println("sum == 0.3: false (正确)")
	}

	// 正确比较方式
	epsilon := 1e-10
	if math.Abs(sum-0.3) < epsilon {
		fmt.Println("使用 epsilon 比较: 相等 (在容差范围内)")
	}
	fmt.Println()
}

// orderOfOperations 演示运算顺序对精度的影响
func orderOfOperations() {
	fmt.Println("--- 先加后乘 vs 先乘后加 ---")

	// 示例 1
	fmt.Println("\n示例 1: 1e-10 + 1e20 - 1e20")
	a := 1e-10
	b := 1e20

	result1 := a + b - b
	result2 := (a - b) + b

	fmt.Printf("先加后减 (a + b - b): %.20e\n", result1)
	fmt.Printf("先减后加 ((a - b) + b): %.20e\n", result2)
	fmt.Printf("期望值: %.20e\n", a)

	// 示例 2: 累加顺序
	fmt.Println("\n示例 2: 累加 0.1 十次")
	sumFloat := 0.0
	for i := 0; i < 10; i++ {
		sumFloat += 0.1
	}
	fmt.Printf("float64 累加结果: %.20f\n", sumFloat)
	fmt.Printf("与 1.0 的差值:    %.20e\n", math.Abs(sumFloat-1.0))

	// 示例 3: 更大数量级的差异
	fmt.Println("\n示例 3: 1e15 + 0.123456789 - 1e15")
	large := 1e15
	small := 0.123456789

	sum1 := large + small - large
	sum2 := (large - large) + small

	fmt.Printf("先加后减: %.15f\n", sum1)
	fmt.Printf("先减后加: %.15f\n", sum2)
	fmt.Printf("期望值:   %.15f\n", small)
	fmt.Println()
}

// largeNumberEatsSmall 演示大数吃小数问题
func largeNumberEatsSmall() {
	fmt.Println("--- 大数吃小数问题 ---")

	x := 1.0
	y := 1e-15

	fmt.Printf("x = %.f, y = %.20f\n", x, y)
	fmt.Printf("x + y - x = %.20f (丢失了 y 的信息)\n", x+y-x)
	fmt.Printf("x - x + y = %.20f (保留了 y)\n", x-x+y)

	// 另一个典型例子
	fmt.Println("\n更极端的例子:")
	fmt.Printf("float64 最大安全整数: %.f\n", float64(1<<53))

	// 超过安全整数范围后精度丢失
	largeInt := float64(1<<53) + 1
	fmt.Printf("1<<53 + 1 = %.f (在安全范围内)\n", largeInt)

	veryLargeInt := float64(1<<53) + 2
	fmt.Printf("1<<53 + 2 = %.f (可能与上面相同，精度丢失)\n", veryLargeInt)

	fmt.Printf("\n差值: %.f\n", veryLargeInt-largeInt)
	fmt.Println()
}

// decimalExample 演示使用 decimal 库解决精度问题
func decimalExample() {
	fmt.Println("--- 使用 decimal 库 ---")

	// 1. 基本运算
	d1 := decimal.NewFromFloat(0.1)
	d2 := decimal.NewFromFloat(0.2)
	sum := d1.Add(d2)
	fmt.Printf("decimal: 0.1 + 0.2 = %s (精确)\n", sum.String())
	fmt.Printf("float64: 0.1 + 0.2 = %.20f (不精确)\n", 0.1+0.2)

	// 2. 金融计算示例
	fmt.Println("\n金融计算示例:")
	price := decimal.NewFromFloat(19.99)
	taxRate := decimal.NewFromFloat(0.08)
	tax := price.Mul(taxRate).Round(2)
	total := price.Add(tax)

	fmt.Printf("商品价格: $%s\n", price.String())
	fmt.Printf("税率(8%%): $%s\n", tax.String())
	fmt.Printf("总价:    $%s\n", total.String())

	// 3. 高精度累加
	fmt.Println("\n累加 0.1 十次:")
	sumDec := decimal.NewFromFloat(0)
	for i := 0; i < 10; i++ {
		sumDec = sumDec.Add(decimal.NewFromFloat(0.1))
	}
	fmt.Printf("decimal 结果: %s\n", sumDec.String())
	fmt.Printf("float64 结果: %.20f\n", 0.1+0.1+0.1+0.1+0.1+0.1+0.1+0.1+0.1+0.1)

	// 4. 解决大数吃小数问题
	fmt.Println("\n解决大数吃小数问题:")
	f1 := 1.0
	f2 := 1e-15

	fmt.Printf("float64: %.20f\n", f1+f2-f1)
	d1f := decimal.NewFromFloat(f1)
	d2f := decimal.NewFromFloat(f2)
	resultDec := d1f.Add(d2f).Sub(d1f)
	fmt.Printf("decimal: %s\n", resultDec.String())

	// 5. 比较操作
	fmt.Println("\n精确比较:")
	d3 := decimal.NewFromFloat(0.1)
	d4 := decimal.NewFromFloat(0.10)
	fmt.Printf("0.1 == 0.10 (decimal): %v\n", d3.Equal(d4))
	fmt.Printf("0.1 == 0.10 (float64): %v\n", 0.1 == 0.10)
}

// 注意: decimal 库需要单独安装
// 运行前请执行: go get github.com/shopspring/decimal
func installDecimalNote() {
	log.Println("安装 decimal 库: go get github.com/shopspring/decimal")
}
