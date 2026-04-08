# 浮点数精度问题

## IEEE 754 表示原理

IEEE 754 是浮点数的国际标准，Go 语言中的 `float32` 和 `float64` 均采用此标准。

### 构成部分

一个浮点数由三部分组成：

| 部分 | float32 | float64 | 说明 |
|------|---------|---------|------|
| 符号位 (Sign) | 1 bit | 1 bit | 0 表示正数，1 表示负数 |
| 指数位 (Exponent) | 8 bits | 11 bits | 存储指数的偏移值 |
| 尾数位 (Mantissa/Fraction) | 23 bits | 52 bits | 存储有效数字 |

### 表示公式

```
(-1)^sign × 1.fraction × 2^(exponent - bias)
```

其中 `bias` 是偏移量，float32 为 127，float64 为 1023。

### 精度问题根源

- **尾数位有限**：float32 只有 23 bits，float64 只有 52 bits
- **十进制数并非都能精确表示**：如 0.1、0.2、0.3 等在二进制中都是无限循环小数
- **数值范围限制**：超出范围的数会被截断或溢出

## 先乘后加 vs 先加后乘的精度差异

在数值计算中，运算顺序会显著影响结果精度。

### 问题演示

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	// 示例：计算 1e-10 + 1e20 - 1e20
	a := 1e-10
	b := 1e20

	// 方式一：先加后减
	result1 := a + b - b
	fmt.Printf("先加后减: %.20f\n", result1)
	fmt.Printf("期望值:   %.20f\n", a)
	fmt.Printf("误差:     %.20e\n", math.Abs(result1-a))

	// 方式二：先减后加
	result2 := (a - b) + b
	fmt.Printf("\n先减后加: %.20f\n", result2)
	fmt.Printf("期望值:   %.20f\n", a)
	fmt.Printf("误差:     %.20e\n", math.Abs(result2-a))
}
```

### 结果分析

```
先加后减: 0.00000000000000000000  // 精度完全丢失
先减后加: 0.00000000000000010000 // 保留了部分精度
```

原因：`b = 1e20` 远大于 `a = 1e-10`，直接 `a + b` 会导致 a 的信息丢失。

### 另一个典型案例

```go
package main

import (
	"fmt"
	"math"
)

func main() {
	// 大数吃小数问题
	x := 1.0
	y := 1e-15

	fmt.Printf("x + y - x = %.20f\n", x+y-x)
	fmt.Printf("x - x + y = %.20f\n", x-x+y)

	// 更极端的例子
	large := 1e15
	small := 0.123456789

	sum1 := large + small - large
	sum2 := (large - large) + small

	fmt.Printf("\n大数: %.f, 小数: %.9f\n", large, small)
	fmt.Printf("先加后减: %.9f\n", sum1)
	fmt.Printf("先减后加: %.9f\n", sum2)
}
```

## 高精度需求推荐使用 decimal 库

对于金融计算、精确计量等场景，推荐使用 `decimal` 库：

### 常用 decimal 库

- ** shopspring/decimal**：广泛使用，API 友好
  - GitHub: https://github.com/shopspring/decimal
  - 安装：`go get github.com/shopspring/decimal`

### 使用示例

```go
package main

import (
	"fmt"
	"log"

	"github.com/shopspring/decimal"
)

func main() {
	// 安装 decimal 库: go get github.com/shopspring/decimal

	// 精确计算：0.1 + 0.2
	d1 := decimal.NewFromFloat(0.1)
	d2 := decimal.NewFromFloat(0.2)
	sum := d1.Add(d2)
	fmt.Printf("0.1 + 0.2 = %s\n", sum.String()) // 输出: 0.3

	// 金融计算示例
	price := decimal.NewFromFloat(19.99)
	taxRate := decimal.NewFromFloat(0.08)
	tax := price.Mul(taxRate).Round(2)
	total := price.Add(tax)
	fmt.Printf("商品价格: %s, 税(8%%): %s, 总价: %s\n",
		price.String(), tax.String(), total.String())

	// 高精度场景
	bigNum := decimal.NewFromFloat(1e10)
	smallNum := decimal.NewFromFloat(0.1)
	result := bigNum.Add(smallNum).Sub(bigNum)
	fmt.Printf("\n1e10 + 0.1 - 1e10 = %s\n", result.String()) // 正确: 0.1

	// 避免精度问题
	// 如果不使用 decimal:
	f1 := 1.0
	f2 := 1e-15
	fmt.Printf("\nfloat64 直接计算: %.20f\n", f1+f2-f1)

	// 使用 decimal:
	d1f := decimal.NewFromFloat(f1)
	d2f := decimal.NewFromFloat(f2)
	resultDec := d1f.Add(d2f).Sub(d1f)
	fmt.Printf("decimal 计算:    %s\n", resultDec.String())
}
```

### decimal 库优势

| 特性 | float64 | decimal.Decimal |
|------|---------|-----------------|
| 精度 | ~15-17 位十进制 | 任意精度（可配置） |
| 性能 | 快 | 较慢 |
| 内存占用 | 低 | 较高 |
| 适用场景 | 科学计算 | 金融计算、精确计量 |
| 0.1 + 0.2 | 0.30000000000000004 | 0.3 |

## 实践建议

1. **避免直接比较浮点数相等**：使用 `math.Abs(a-b) < epsilon` 而非 `a == b`

2. **运算顺序影响结果**：尽量先加减小数值，再处理大数值

3. **金融计算必须用 decimal**：任何涉及金钱的计算都不能使用 float64

4. **注意数值溢出**：
   ```go
   // float64 最大值约 1.8e308
   fmt.Println(math.MaxFloat64) // 1.7976931348623157e+308
   ```

5. **类型转换要小心**：
   ```go
   var f float64 = 1.0000000000000001
   var i int = int(f) // 结果可能是 1，不是预期的精确转换
   ```

## 延伸阅读

- [IEEE 754 维基百科](https://en.wikipedia.org/wiki/IEEE_754)
- [golang decimal 库](https://github.com/shopspring/decimal)
- [浮点数端到端解析](https://float.exposed/)
