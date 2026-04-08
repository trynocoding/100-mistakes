# 逃逸分析（Escape Analysis）

## 概念

逃逸分析是 Go 编译器自动分析代码，决定变量应该分配在栈上还是堆上的过程。分配在栈上的变量在函数返回后自动回收，而分配在堆上的变量需要垃圾回收器（GC）来处理。

Go 编译器会尽可能将变量分配在栈上，以减少 GC 的压力，提高程序性能。

## 栈分配 vs 堆分配

| 特性 | 栈分配 | 堆分配 |
|------|--------|--------|
| 内存位置 | 栈（Stack） | 堆（Heap） |
| 分配速度 | 极快（仅移动栈指针） | 较慢（需要 GC） |
| 回收方式 | 函数返回时自动回收 | 垃圾回收器回收 |
| 生命周期 | 作用域内 | 由垃圾回收器管理 |
| 性能开销 | 低 | 高（GC 暂停） |

## 查看逃逸分析结果

使用 `-gcflags '-m -l'` 查看逃逸分析信息：

```bash
go run -gcflags '-m -l' main.go
```

- `-m`：显示逃逸分析决策
- `-l`：禁用内联（以便看到更清晰的分析结果）

### 示例输出

```
./main.go:10:2: leaking param: x
./main.go:15:2: y escapes to heap
./main.go:20:2: z does not escape
```

## 常见逃逸场景

### 1. 返回指针

当函数返回局部变量的指针时，变量会逃逸到堆上。

```go
func bad() *int {
    v := 10      // v 分配在栈上
    return &v    // 错误：v 逃逸到堆上，因为返回了指针
}
```

正确做法：让调用者决定内存分配

```go
func good(result *int) {
    *result = 10 // 将指针作为参数传入，在调用者分配的内存上写入
}
```

### 2. 接口类型（any）返回

当返回 `any`（空接口）类型时，指针指向的值会逃逸到堆上，因为接口需要存储具体的类型信息。

```go
func interfaceEscape() any {
    v := 10
    return v // v 逃逸到堆上
}
```

### 3. 闭包捕获

闭包引用外部变量时，该变量会逃逸到堆上。

```go
func closureEscape() func() int {
    v := 10
    return func() int { // v 逃逸到堆上，被闭包引用
        return v
    }
}
```

### 4. 切片长度/容量为变量

当切片的长度和容量取决于运行时才能确定的变量时，切片可能会逃逸。

```go
func sliceEscape(n int) []int {
    // 如果 n 是编译期常量，编译器可能优化为栈分配
    // 但如果是变量，Go 1.21+ 会尽量在栈上分配
    return make([]int, n)
}
```

### 5. 不确定大小的数组

```go
func stackVsHeap(n int) {
    // 编译期不确定大小，分配在堆上
    arr := make([]int, n)
    _ = arr
}
```

## 验证方法

### 方法 1：使用 go run 查看分析

```bash
go run -gcflags '-m -l' code.go
```

### 方法 2：使用 go build 查看

```bash
go build -gcflags '-m -l' -o /dev/null code.go
```

### 方法 3：在代码中嵌入注释

在变量声明旁添加 `//go:embed` 或使用 `fmt.Println` 强制变量存在

### 方法 4：使用 pprof 查看堆分配

```bash
go test -bench=. -benchmem -memprofile=mem.prof
go tool pprof mem.prof
```

## 最佳实践

1. **避免不必要的指针传递**：只在必要时使用指针
2. **优先使用栈分配**：让编译器做优化
3. **注意接口的使用**：返回具体类型而非 `any`
4. **谨慎使用闭包**：闭包会导致捕获的变量逃逸
5. **使用 sync.Pool 复用对象**：减少堆分配频率

## 常见误区的澄清

1. **Go 不会自动将所有变量放到堆上**：编译器会进行严格的逃逸分析
2. **返回局部变量指针不一定是错的**：如果正确使用，Go 可以处理
3. **闭包不一定会逃逸**：只有引用了外部变量时才会逃逸

## 参考

- [Go 官方博客：Go escape analysis](https://go.dev/blog/escape)
- [Go 命令行文档：gcflags](https://pkg.go.dev/cmd/compile)
- [Go Wiki: Escape Analysis](https://github.com/golang/go/wiki/EscapeAnalysis)