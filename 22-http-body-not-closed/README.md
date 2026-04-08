# HTTP Body 未关闭

## 知识点描述

HTTP 响应体（`resp.Body`）必须显式关闭，以释放底层网络连接资源。在 Go 中，HTTP 客户端复用 TCP 连接，如果不关闭响应体，会导致连接泄漏，最终可能耗尽文件描述符或连接池。

## 关键概念

### 1. defer resp.Body.Close() 的正确用法

关闭 HTTP 响应体的标准方式是在处理响应后立即使用 `defer` 关闭。

```go
resp, err := http.Get("https://example.com")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close() // 确保函数返回前关闭
```

### 2. 使用 io.ReadAll 或 iooutil.ReadAll 后仍需关闭

很多开发者误以为读取完响应体内容后就不需要关闭了，这是一个常见的错误。即使已经读取了所有内容，响应体仍然需要关闭以释放连接资源。

```go
// 错误示例：读取了内容但没有关闭
body, _ := io.ReadAll(resp.Body)
fmt.Println(string(body))
// resp.Body 未关闭，连接泄漏！

// 正确示例
body, _ := io.ReadAll(resp.Body)
fmt.Println(string(body))
resp.Body.Close() // 仍然需要关闭
```

### 3. Go 1.23+ 的简化方式

Go 1.23 引入了 `http.MethodDo` 系列方法的改进，`client.Do` 方法现在可以自动处理响应体的关闭。但需要注意，这仅适用于使用 `http.Client` 的场景。

```go
// Go 1.23+：可以使用更简洁的方式
resp, err := client.Do(req)
// 不再需要手动 defer resp.Body.Close()
// Go 1.23+ 自动处理了响应体的关闭
```

不过，为了代码的兼容性和可读性，建议始终显式关闭响应体，除非明确知道运行环境是 Go 1.23+。

## 常见错误示例

```go
// 错误 1：忘记关闭
func badGet(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    // 忘记关闭 resp.Body
    body, err := io.ReadAll(resp.Body)
    return string(body), err
}

// 错误 2： defer 位置错误
func deferredTooLate(url string) error {
    resp, err := http.Get(url)
    defer resp.Body.Close() // 正确位置
    if err != nil {
        return err
    }
    // defer 会在函数结束时执行，正确
    body, err := io.ReadAll(resp.Body)
    return err
}
```

## 正确的做法总结

| 场景 | 推荐做法 |
|------|----------|
| Go 1.21-1.22 | 必须使用 `defer resp.Body.Close()` |
| Go 1.23+ | `client.Do` 自动关闭，但仍建议显式关闭以确保兼容性 |

## 相关知识点

- HTTP 连接复用机制
- 文件描述符泄漏
- Go 1.23 HTTP 客户端改进
- `io.ReadAll` vs `io.Copy` 性能比较

## 代码示例

请参考 [code.go](./code.go) 获取完整的可运行示例代码。