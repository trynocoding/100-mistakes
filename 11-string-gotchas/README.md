# String 处理注意事项

## 概述

Go 语言中 string 是 UTF-8 编码的字节序列。理解 UTF-8 的工作原理对于正确处理字符串至关重要。本章节将介绍几个常见的 string 处理陷阱。

---

## 陷阱一：字节索引 vs 字符索引

### 问题描述

UTF-8 编码中，不同字符占据的字节数不同：
- ASCII 字符（英文字母、数字、标点）：1 字节
- 中文、日文、韩文等：3 字节
- 带附加符号的字符：2 字节以上

如果使用字节索引访问字符串，可能无法获取完整的字符。

### 问题演示

```go
s := "Hello世界"

fmt.Println("字符串长度（字节）:", len(s))                    // 13
fmt.Println("字符串长度（字符）:", utf8.RuneCountInString(s)) // 9

// 错误：按字节索引访问
fmt.Printf("s[6] = %c (可能是乱码或半个字符)\n", s[6])

// 正确：按字符遍历
for i, r := range s {
    fmt.Printf("s[%d] = %c\n", i, r)
}
```

### 正确做法

**始终使用 `for i, r := range s` 遍历 Unicode 字符**：

```go
s := "Hello世界"
for i, r := range s {
    fmt.Printf("索引 %d 处的字符: %c\n", i, r)
}
```

如果需要将字符串转换为字符切片：

```go
s := "Hello世界"
runes := []rune(s) // 转换为 rune 切片
fmt.Println(string(runes[6])) // 正确输出: 界
```

---

## 陷阱二：切片截断与内存泄漏

### 问题描述

当对 string 进行切片操作时（如 `s[start:end]`），新字符串会引用原始字符串的底层字节数组。即使你只使用了很小的一部分，原始的大字符串仍然无法被垃圾回收。

### 问题演示

```go
func main() {
    large := make([]byte, 1024*1024) // 1MB 数据
    for i := range large {
        large[i] = 'A'
    }
    s := string(large) // 创建字符串

    small := s[:10] // 只使用前 10 个字节

    // 问题：即使 small 很小，large 仍然无法被回收
    // 因为 small 引用了 large 的底层数组

    runtime.GC()
    fmt.Println("small:", small)
}
```

### 正确做法

**使用 `strings.Clone` 或 `string(b)` 创建独立副本**：

```go
small := s[:10]
cloned := strings.Clone(small) // 独立副本，可以释放原字符串
// 或者
cloned := string([]byte(small)) // 强制创建新字符串
```

### 内存泄漏场景

| 操作 | 风险 | 解决方案 |
|------|------|----------|
| `s[start:end]` 切片大字符串 | 引用原始数据，无法回收 | `strings.Clone()` |
| 将大字符串赋值给结构体字段 | 引用原始数据 | `string([]byte(s))` |
| 在映射中使用字符串键 | 引用原始数据 | 确保键短小或克隆 |

---

## 陷阱三：strings.Clone vs 手动复制

### 问题描述

在 Go 1.18 之前，复制字符串需要 `string([]byte(s))` 这种方式。Go 1.18 引入了 `strings.Clone`，它是一种更高效的实现。

### 性能对比

```go
import "strings"

s := "Hello, 世界！这是一个很长的字符串..."

// 方式一：手动复制（两次内存分配）
s1 := string([]byte(s))

// 方式二：strings.Clone（一次内存分配）
s2 := strings.Clone(s)

// strings.Clone 在以下情况下更优：
// 1. 字符串较长
// 2. 需要避免临时 []byte 分配
```

### strings.Clone 的优势

```go
// strings.Clone 内部实现
func Clone(s string) string {
// 如果字符串为空，直接返回
if len(s) == 0 {
    return ""
}
// 分配新字符串并复制
b := make([]byte, len(s))
copy(b, s)
return string(b)
}
```

对比传统方式 `string([]byte(s))`：
- 减少一次内存分配
- 避免 `[]byte` 临时对象创建
- 代码意图更清晰

### 正确做法

```go
// Go 1.18+ 推荐使用
cloned := strings.Clone(original)

// 旧版本兼容写法
cloned := string([]byte(original))
```

---

## 陷阱四：字符串拼接的常见误区

### 问题描述

频繁拼接字符串时，使用 `+` 操作符会创建大量临时字符串，影响性能。

### 正确做法

使用 `strings.Builder` 或 `strings.Join`：

```go
// 低效：每次 + 都创建新字符串
var result string
for _, s := range strs {
    result += s
}

// 高效：使用 strings.Builder
var b strings.Builder
for _, s := range strs {
    b.WriteString(s)
}
result := b.String()

// 或者使用 strings.Join（最简洁）
result := strings.Join(strs, "")
```

---

## 最佳实践总结

| 场景 | 错误做法 | 正确做法 |
|------|----------|----------|
| 遍历 Unicode 字符 | `for i := 0; i < len(s); i++` | `for i, r := range s` |
| 访问字符串元素 | `s[i]` | `[]rune(s)[i]` 或 `utf8.DecodeRuneInString(s[i:])` |
| 切片大字符串 | `small := large[0:10]` | `small := strings.Clone(large[0:10])` |
| 复制字符串 | `string([]byte(s))` | `strings.Clone(s)` (Go 1.18+) |
| 拼接多个字符串 | `s1 + s2 + s3` | `strings.Builder` 或 `strings.Join` |

---

## 参考

- [Go 语言规范 - 字符串字面量](https://go.dev/ref/spec#String_literals)
- [Go Blog - Strings, bytes, runes and characters in Go](https://go.dev/blog/strings)
- [strings.Clone 文档](https://pkg.go.dev/strings#Clone)
- [unicode/utf8 包](https://pkg.go.dev/unicode/utf8)
