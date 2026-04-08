# Defer 注意事项

## 概述

`defer` 是 Go 语言中一种延迟执行的机制，常用于资源释放、解锁互斥锁、关闭文件等场景。然而，`defer` 有一些容易忽略的细节和陷阱，本文将通过代码演示三个关键注意点。

---

## 1. 函数永不返回时，defer 永不执行

当一个函数永远不会返回时（例如进入死循环），该函数中注册的 `defer` 语句将**永远不会**被执行。

### 问题演示

```go
func exampleNeverReturns() {
    defer fmt.Println("This will never be printed") // 不会执行

    for {
        time.Sleep(time.Hour) // 死循环，函数永不返回
    }
}
```

### 关于 runtime.Goexit()

`runtime.Goexit()` 的行为比较特殊：它会终止当前 goroutine，但在终止前会执行所有**已经注册**的 defer。这意味着：

- `Goexit()` 调用之前注册的 defer 会执行
- `Goexit()` 之后的代码永远不会执行

```go
func exampleGoexit() {
    defer fmt.Println("This WILL be printed") // 会执行

    runtime.Goexit()
    fmt.Println("This will NEVER be printed") // 不会执行
}
```

### 关键点

- `defer` 只能在函数**正常返回**、通过 `panic` 导致栈展开时、或 `Goexit()` 终止 goroutine 时才会执行
- 死循环会导致函数永不返回，函数内的 `defer` 永不执行
- `runtime.Goexit()` 会执行已注册的 defer，然后立即终止 goroutine

---

## 2. defer 参数求值时机：声明时计算

`defer` 语句中使用的参数，**不是在 defer 执行时求值，而是在 defer 声明时求值**。这是 Go新手常犯的错误之一。

### 问题演示

```go
func deferredParameterEvaluation() {
    x := 1
    defer fmt.Println("Deferred:", x) // x 在声明时的值是 1

    x = 2
    fmt.Println("Before return:", x) // 输出 2
    // 输出: Before return: 2
    // 函数返回后 defer 执行，输出: Deferred: 1
}
```

### 正确做法

如果需要使用函数返回时的变量值，应使用**闭包**：

```go
func deferredClosureEvaluation() {
    x := 1
    defer func() {
        fmt.Println("Deferred:", x) // 闭包捕获 x 的引用
    }()

    x = 2
    fmt.Println("Before return:", x) // 输出 2
    // 函数返回后 defer 执行，闭包中的 x 已经是 2，输出: Deferred: 2
}
```

### 关键点

| 写法 | 行为 |
|------|------|
| `defer fmt.Println(x)` | x 在 defer 声明时求值 |
| `defer func() { fmt.Println(x) }()` | 闭包在 defer 执行时求值，捕获变量引用 |

---

## 3. 闭包中捕获变量的不同行为

闭包捕获变量时有两种方式：**按值捕获**和**按引用捕获**。Go 中闭包总是捕获变量的引用，而非值。这意味着闭包内部对变量的修改会影响外部变量。

### 问题演示

```go
func closureVariableCapture() {
    funcs := make([]func(), 3)
    for i := 0; i < 3; i++ {
        funcs[i] = func() {
            fmt.Println("Value:", i) // 所有闭包都捕获同一个 i
        }
    }

    for _, f := range funcs {
        f() // 输出: Value: 3, Value: 3, Value: 3
    }
}
```

### 正确做法

通过函数参数传递值，或创建局部变量：

```go
func closureCorrectCapture() {
    funcs := make([]func(), 3)
    for i := 0; i < 3; i++ {
        // 方法 1：通过函数参数按值传递
        funcs[i] = func(val int) func() {
            return func() {
                fmt.Println("Value:", val)
            }
        }(i)

        // 方法 2：创建局部变量
        // i := i
        // funcs[i] = func() { fmt.Println("Value:", i) }
    }

    for _, f := range funcs {
        f() // 输出: Value: 0, Value: 1, Value: 2
    }
}
```

### 关键点

- Go 闭包捕获的是变量的**引用**，而非值
- 在循环中创建闭包时，需要注意变量的生命周期
- 正确的做法是使用函数参数传递值，或创建新的局部变量来"快照"当前值

---

## 代码示例

完整的可运行代码请参见 [code.go](./code.go)。运行方式：

```bash
go run code.go
```

---

## 总结

| 陷阱 | 原因 | 解决方案 |
|------|------|----------|
| defer 永不执行 | 函数永不返回（Goexit/死循环） | 将 defer 放在会返回的函数中 |
| 参数值不对 | defer 参数在声明时求值 | 使用闭包捕获变量引用 |
| 闭包捕获错误 | 循环中闭包共享同一变量引用 | 通过函数参数传递或创建局部变量 |
