# 字符串与字节切片转换

## 概述

Go 语言中，字符串（string）和字节切片（[]byte）是两种不同的类型。字符串是不可变的 UTF-8 字节序列，而字节切片是可变的底层数组。理解它们之间的转换及其内存行为对于编写高效的 Go 程序至关重要。

---

## 内存分配问题

### 传统转换方式

在 Go 中，传统方式使用类型转换 `[]byte(s)` 和 `string(b)`：

```go
s := "hello"
b := []byte(s)  // 分配新字节切片，复制数据
s2 := string(b) // 分配新字符串，复制数据
```

**问题**：每次转换都会分配新的内存并复制数据。对于频繁转换的场景，这会造成不必要的内存分配和 GC 压力。

### 内存分配对比

| 转换方式 | 内存分配 | 数据复制 | 是否安全 |
|----------|----------|----------|----------|
| `[]byte(s)` | 是 | 是 | 安全 |
| `string(b)` | 是 | 是 | 安全 |
| `unsafe` 方式 (旧) | 否 | 否 | 有风险 |
| `unsafe.String()` (Go 1.15+) | 否 | 否 | 安全但需谨慎 |

---

## 旧版 unsafe 方式（Go 1.12-1.14）

### 实现方式

在 Go 1.12-1.14 中，社区广泛使用以下方式避免内存分配：

```go
// 字符串转字节切片（避免分配）
func StringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}

// 字节切片转字符串（避免分配）
func BytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}
```

### 严重问题

**问题一：违反 Go 内存模型**

字符串被设计为不可变，但通过这种转换得到的字节切片是可变的。对切片的后续修改可能引发未定义行为：

```go
s := "hello"
// 通过 unsafe 转换得到可变切片
b := StringToBytes(s)
b[0] = 'H' // 危险！修改了字符串的底层数据
fmt.Println(s) // 可能输出 "Hello"（取决于 Go 实现）
```

**问题二：依赖内部结构**

这种实现直接依赖 `string` 和 `slice` 的内部结构：
- `string` 结构：`{ ptr *byte, len int }`
- `slice` 结构：`{ ptr *byte, len int, cap int }`

如果 Go 运行时改变这些结构，代码将崩溃。

**问题三：切片操作导致内存泄漏**

```go
s := strings.Repeat("A", 100)
b := StringToBytes(s)
small := b[:10] // small 引用整个原始字符串的底层数组
// 即使只使用 10 字节，100 字节的字符串仍无法被 GC 回收
```

---

## 现代 Go API（Go 1.15+）

### unsafe.String()

Go 1.15 引入了 `unsafe.String()`，将字节指针转换为字符串：

```go
// 函数签名
func String(ptr *byte, len IntegerType) string
```

**特性**：
- 不分配新内存
- 字符串直接引用传入的字节内存
- 字符串不可变（语义层面）

### unsafe.Slice()

Go 1.17 引入了 `unsafe.Slice()`，将指针转换为切片：

```go
// 函数签名
func Slice(ptr *byte, len IntegerType) []byte
```

### 现代推荐方式

```go
import "unsafe"

func StringToBytes(s string) []byte {
    if s == "" {
        return nil
    }
    return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BytesToString(b []byte) string {
    if len(b) == 0 {
        return ""
    }
    return unsafe.String(&b[0], len(b))
}
```

### 优势

1. **官方支持**：`unsafe.String()` 和 `unsafe.Slice()` 是官方提供的 API
2. **明确语义**：文档清楚说明了使用时的注意事项
3. **编译期检查**：配合 `go vet` 可以发现一些潜在问题

### 注意事项

使用这些 API 时必须遵循以下规则：

**规则一：字节数据必须来自可寻址的数组**

```go
// 错误：无法获取栈上数组的指针
arr := [5]byte{'h', 'e', 'l', 'l', 'o'}
s := unsafe.String(&arr[0], 5) // 编译错误

// 正确：使用堆上的切片
arr := []byte{'h', 'e', 'l', 'l', 'o'}
s := unsafe.String(&arr[0], len(arr))
```

**规则二：不要修改通过 `unsafe.String()` 创建的字符串对应底层字节**

```go
s := "hello"
b := unsafe.Slice(unsafe.StringData(s), len(s))
b[0] = 'H' // 危险！违反字符串不可变性
```

**规则三：注意字符串生命周期**

```go
func bad() []byte {
    s := "hello"
    b := unsafe.Slice(unsafe.StringData(s), len(s))
    return b // 危险！s 在函数返回后生命周期结束
}
```

---

## 性能对比

### 内存分配对比

| 方法 | 分配次数 | 复制数据 | 适用场景 |
|------|----------|----------|----------|
| `[]byte(s)` | 1 次 | 是 | 临时使用，需要修改 |
| `string(b)` | 1 次 | 是 | 临时使用，需要修改 |
| `unsafe` 旧方式 | 0 次 | 否 | **不推荐** |
| `unsafe` 现代方式 | 0 次 | 否 | 高性能场景，需确保安全 |

### 何时使用 unsafe 转换

**适合的场景**：
1. 高频调用路径，对性能敏感
2. 确认数据来源可靠，不会被 GC 回收
3. 转换后的字节切片不会被修改

**不适合的场景**：
1. 一般业务代码
2. 需要修改数据时
3. 数据来源不确定时

---

## 最佳实践

### 性能要求不高的场景

```go
// 直接使用传统方式，代码清晰安全
s := "hello"
b := []byte(s)  // 简单明了
s2 := string(b)
```

### 高性能场景

```go
import "unsafe"

// 使用包级变量缓存空字符串和空切片
var emptyString = ""
var emptyBytes = []byte{}

func StringToBytes(s string) []byte {
    if s == "" {
        return emptyBytes
    }
    return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BytesToString(b []byte) string {
    if len(b) == 0 {
        return emptyString
    }
    return unsafe.String(&b[0], len(b))
}
```

### 使用 bytes.Clone()（Go 1.20+）

如果需要创建独立副本，避免共享底层数组：

```go
import "bytes"

s := "hello"
b := []byte(s)
bClone := bytes.Clone(b) // 创建独立副本，避免修改影响原字符串
```

---

## 总结

| 场景 | 推荐方式 | 原因 |
|------|----------|------|
| 一般转换 | `[]byte(s)` / `string(b)` | 安全、清晰 |
| 高性能、确定安全 | `unsafe.Slice()` / `unsafe.String()` | 避免分配 |
| 需要独立副本 | `bytes.Clone()` / `strings.Clone()` | 安全复制 |
| 旧版 unsafe 方式 | **禁止使用** | 危险、依赖内部实现 |

**核心原则**：除非有明确的性能需求且确认安全，否则应使用传统转换方式。现代 `unsafe` API 虽然比旧版安全，但仍然需要谨慎使用。

---

## 参考

- [Go 官方文档 - unsafe 包](https://pkg.go.dev/unsafe)
- [Go Wiki - Strings](https://github.com/golang/go/wiki/Strings)
- [Go Blog - The Go Programming Language Blog: Strings](https://go.dev/blog/strings)
- [Go Issue #25471 - proposal: add unsafe.String](https://github.com/golang/go/issues/25471)
