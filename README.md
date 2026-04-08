# 100 Go Mistakes - Go语言100个常见错误

这是一份关于Go语言常见错误的知识手册，内容源自《100 Go Mistakes》并针对现代Go版本（1.21+）进行了重写和验证。

## 内容列表

### 基础语法与概念
| 编号 | 知识点 | 目录 |
|------|--------|------|
| 01 | 变量遮蔽 (Variable Shadowing) | [01-variable-shadowing/](01-variable-shadowing/) |
| 02 | init函数使用注意事项 | [02-init-function/](02-init-function/) |
| 03 | 嵌入类型 (Embedded Types) | [03-embedded-types/](03-embedded-types/) |
| 04 | 函数式选项模式 (Functional Options) | [04-functional-options/](04-functional-options/) |
| 05 | 八进制整数陷阱 | [05-octal-integers/](05-octal-integers/) |
| 06 | 浮点数精度问题 | [06-floating-point/](06-floating-point/) |

### 数据结构
| 编号 | 知识点 | 目录 |
|------|--------|------|
| 07 | Slice注意事项 | [07-slice-gotchas/](07-slice-gotchas/) |
| 08 | Range注意事项 | [08-range-gotchas/](08-range-gotchas/) |
| 09 | Break作用域问题 | [09-break-scope/](09-break-scope/) |
| 10 | Defer注意事项 | [10-defer-gotchas/](10-defer-gotchas/) |
| 11 | String处理注意事项 | [11-string-gotchas/](11-string-gotchas/) |
| 12 | 接口返回非nil问题 | [12-interface-nil/](12-interface-nil/) |

### 错误处理与并发
| 编号 | 知识点 | 目录 |
|------|--------|------|
| 13 | Error处理 | [13-error-handling/](13-error-handling/) |
| 14 | Happens-Before保证 | [14-happens-before/](14-happens-before/) |
| 15 | Context Values | [15-context-values/](15-context-values/) |
| 16 | Goroutine生命周期管理 | [16-goroutine-lifecycle/](16-goroutine-lifecycle/) |
| 17 | Channel注意事项 | [17-channel-gotchas/](17-channel-gotchas/) |
| 18 | String()方法死锁 | [18-stringer-deadlock/](18-stringer-deadlock/) |
| 19 | WaitGroup正确用法 | [19-waitgroup-usage/](19-waitgroup-usage/) |
| 20 | 不要拷贝sync类型 | [20-sync-copy-gotcha/](20-sync-copy-gotcha/) |

### 性能与优化
| 编号 | 知识点 | 目录 |
|------|--------|------|
| 21 | time.After内存泄漏 | [21-time-after-leak/](21-time-after-leak/) |
| 22 | HTTP Body未关闭 | [22-http-body-not-closed/](22-http-body-not-closed/) |
| 23 | Cache Line优化 | [23-cache-line-optimization/](23-cache-line-optimization/) |
| 24 | False Sharing | [24-false-sharing/](24-false-sharing/) |
| 25 | 内存对齐 | [25-memory-alignment/](25-memory-alignment/) |
| 26 | 逃逸分析 | [26-escape-analysis/](26-escape-analysis/) |
| 27 | 字符串与字节切片转换 | [27-string-byte-conversion/](27-string-byte-conversion/) |
| 28 | 容器中的GOMAXPROCS | [28-container-gomaxprocs/](28-container-gomaxprocs/) |

## 每个知识点的结构

每个目录包含：
- **README.md** - 详细的知识点说明（中文）
- **code.go** - 可运行的Go代码示例

## 运行代码

```bash
# 进入对应目录
cd 01-variable-shadowing

# 运行示例代码
go run code.go
```

## 贡献

欢迎提交Issue和Pull Request来改进内容！

## License

MIT
