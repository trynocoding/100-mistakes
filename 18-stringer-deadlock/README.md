# Stringer 死锁问题

## 概述

在 Go 语言中，实现了 `Stringer` 接口的类型可以通过 `fmt.Sprintf` 等函数自动调用其 `String()` 方法来获取字符串表示。然而，如果在 `fmt.Errorf` 的格式化字符串中直接调用自身类型的 `String()` 方法，可能导致死锁。

这是一个隐蔽的并发问题，因为代码看起来完全正常，但在特定条件下会触发死锁。

---

## 问题分析

### 问题代码示例

```go
type User struct {
    ID   int
    Name string
    mu   sync.Mutex
}

func (u *User) String() string {
    u.mu.Lock()
    defer u.mu.Unlock()
    return fmt.Sprintf("User{ID: %d, Name: %s}", u.ID, u.Name)
}

// 在 fmt.Errorf 中调用 String() 方法
err := fmt.Errorf("用户信息: %s", user.String())
```

### 死锁原因

当 `String()` 方法内部持有自己的锁，并调用 `fmt.Sprintf` 或 `fmt.Errorf` 时，如果这些格式化函数需要通过 `Stringer` 接口（即使用 `%s` 格式化）调用 `String()` 方法，就会触发死锁。

1. `LogError()` 获取 `user.mu` 锁
2. `LogError()` 调用 `fmt.Sprintf("... %s ...", u.String())`
3. `fmt.Sprintf` 内部需要格式化 `u.String()`
4. `u.String()` 尝试获取 `user.mu` 锁
5. 由于 `user.mu` 已被 `LogError()` 持有，`String()` 阻塞等待
6. `LogError()` 阻塞等待 `fmt.Sprintf` 完成
7. **死锁**：两个操作都在等待对方释放资源

关键点：`sync.Mutex` 是不可重入的，同一个 goroutine 不能多次获取同一把锁。

### 何时触发

这个问题在以下场景容易触发：
- 同一个 `User` 对象被多个 goroutine 同时访问
- 其中一个 goroutine 在持有 `User.mu` 锁时执行了日志记录或错误包装操作
- 另一个 goroutine 尝试调用 `fmt.Errorf` 并间接调用 `String()`

---

## 正确处理方式

### 方式一：避免在持有锁时调用 fmt 相关函数

```go
func (u *User) String() string {
    u.mu.Lock()
    defer u.mu.Unlock()
    // 不再调用 fmt.Errorf，而是直接返回字符串
    return fmt.Sprintf("User{ID: %d, Name: %s}", u.ID, u.Name)
}
```

### 方式二：使用 fmt.Stringer 接口而非自定义 String() 方法

```go
type User struct {
    ID   int
    Name string
    mu   sync.Mutex
}

// 分离 Stringer 接口和业务逻辑
func (u *User) GetFormatted() string {
    u.mu.Lock()
    defer u.mu.Unlock()
    return fmt.Sprintf("User{ID: %d, Name: %s}", u.ID, u.Name)
}
```

### 方式三：先获取数据，再格式化

```go
func (u *User) Error() string {
    u.mu.Lock()
    id := u.ID
    name := u.Name
    u.mu.Unlock()
    // 在锁外进行格式化
    return fmt.Sprintf("User{ID: %d, Name: %s}", id, name)
}
```

### 方式四：使用 defer 延迟解锁

```go
func (u *User) String() string {
    u.mu.Lock()
    defer u.mu.Unlock()
    // 如果需要调用 fmt 家族函数，使用延迟调用
    // 但要确保不会再次调用自身
    return u.getStringUnsafe()
}
```

---

## 关键教训

1. **避免在持有锁时调用未知函数**：尤其是 `fmt.Printf`、`fmt.Errorf`、`log.Printf` 等可能间接调用 `String()` 方法的函数。

2. **最小化锁内的操作**：只在锁内进行必要的读/写操作，将格式化等操作移到锁外。

3. **使用值Receiver而非指针Receiver**：如果可以的话，使用值 receiver 可以避免一些锁竞争问题，但这取决于你的业务逻辑是否需要指针。

4. **代码审查时注意**：审查并发代码时，注意 `String()`、`Error()` 等方法内部是否调用了可能触发锁竞争的操作。

---

## 完整代码

请参见 [code.go](./code.go) 文件，该文件包含死锁场景的演示和正确处理方式的示例。

### 代码说明

- `badStringerDeadlock`: 演示会导致死锁的错误写法
- `goodStringer`: 演示正确的 Stringer 实现方式
- `separateLockAndFormat`: 演示先获取数据再格式化的正确方式

### 运行代码

```bash
go run code.go
```

> **注意**：运行代码会触发死锁，这是预期行为，用于演示问题。在真实场景中应避免这种写法。
