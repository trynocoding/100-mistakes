# Break 作用域问题

## 概念

在 Go 语言中，`break` 语句默认只会跳出最近的 `switch` 或 `select` 语句，**不会跳出外层的 `for` 循环**。这是一个容易出错的细节，可能导致程序逻辑与预期不符。

## 错误示例

### 示例：break 只跳出 switch，不跳出 for 循环

```go
func findValue(matrix [][]int, target int) bool {
    for _, row := range matrix {
        for _, col := range row {
            switch col {
            case target:
                // 错误：这里 break 只会跳出 switch，不会跳出外层 for 循环
                // 程序会继续执行外层循环的下一次迭代
                break
            }
            fmt.Println("检查:", col)
        }
    }
    return false
}
```

上述代码的问题在于：当找到目标值时，`break` 语句只跳出了 `switch`，但外层的 `for` 循环继续执行，导致可能继续处理不必要的数据。

## 正确写法

### 方法 1：使用 return 代替 break

```go
func findValueFixed1(matrix [][]int, target int) bool {
    for _, row := range matrix {
        for _, col := range row {
            switch col {
            case target:
                return true // 正确：直接返回，跳出所有循环
            }
            fmt.Println("检查:", col)
        }
    }
    return false
}
```

### 方法 2：使用 label 跳出外层循环

```go
func findValueFixed2(matrix [][]int, target int) bool {
loop:
    for _, row := range matrix {
        for _, col := range row {
            switch col {
            case target:
                break loop // 正确：使用 label 跳出外层循环
            }
            fmt.Println("检查:", col)
        }
    }
    return false
}
```

## Label 的正确用法

### Label 的语法规则

1. Label 名称后必须紧跟一个语句（不能单独占一行）
2. `break label` 会跳转到 label 所在的外层语句块之后执行
3. Label 可以用于 `for`、`switch`、`select` 语句

### Label 跳出多重循环

```go
// 错误写法：break 无法跳出外层循环
func badExample() {
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            switch {
            case i*j > 4:
                break // 只跳出 switch，外层 for 继续执行
            }
        }
    }
}

// 正确写法：使用 label
func goodExample() {
loop:
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            switch {
            case i*j > 4:
                break loop // 跳出外层 for 循环
            }
        }
    }
}
```

### Label 在嵌套 switch 中的使用

```go
func nestedSwitchExample(n int) {
loop:
    for i := 0; i < n; i++ {
        switch i {
        case 2:
            for j := 0; j < n; j++ {
                switch j {
                case 1:
                    // 错误：如果用普通 break，只会跳出内层 switch
                    // 使用 break loop 跳出外层 for 循环
                    break loop
                }
            }
        }
    }
}
```

## 最佳实践

1. **优先使用 return**：如果可以重构为使用 `return`，这是最简洁的方式
2. **使用 label 跳出多重循环**：当需要从嵌套循环中间退出时，使用 label
3. **避免过度嵌套**：过多的嵌套层次会使代码难以阅读和维护，考虑重构
4. **注释说明**：如果必须使用 label 跳出循环，添加注释说明意图

## 常见场景

| 场景 | 错误写法 | 正确写法 |
|------|---------|---------|
| 在 for 的 switch 中找到目标 | `break` | `return` 或 `break label` |
| 嵌套循环中查找元素 | `break` | `break label` |
| 条件满足时停止处理 | `break` (仅跳出 switch) | `break loop` 或 `return` |

## 参考

- [Go 语言规范: Break statements](https://go.dev/ref/spec#Break_statements)
- [Go 语言规范: Labeled statements](https://go.dev/ref/spec#Labeled_statements)