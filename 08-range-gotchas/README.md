# Range 注意事项

## 概述

`range` 是 Go 中用于遍历数组、切片、映射和通道的关键字。但其行为有一些容易忽略的陷阱，理解这些陷阱对于编写正确的 Go 代码至关重要。

---

## 陷阱一：值拷贝问题

### 问题描述

在使用 `range` 遍历切片时，**遍历的值是切片元素的拷贝，而不是元素本身**。因此，修改这个拷贝不会影响原切片。

### 问题演示

```go
nums := []int{1, 2, 3}

for i, n := range nums {
    n = n * 2 // 修改的是拷贝，不影响原切片
}

fmt.Println(nums) // 输出: [1 2 3]（未改变）
```

### 正确做法

如果要修改原切片中的元素，**必须使用索引**：

```go
nums := []int{1, 2, 3}

for i := range nums {
    nums[i] = nums[i] * 2
}

fmt.Println(nums) // 输出: [2 4 6]
```

---

## 陷阱二：指针别名问题

### 问题描述

当遍历切片时（Go 1.21 及更早版本），如果每次迭代都创建一个指针，并将其收集到一个切片中，会导致所有指针都指向同一个地址（最后一次迭代的值）。

### 问题演示（Go 1.21 及更早版本）

```go
nums := []int{1, 2, 3}
var pointers []*int

for _, n := range nums {
    pointers = append(pointers, &n)
}

fmt.Println(*pointers[0], *pointers[1], *pointers[2]) // Go 1.21-: 3 3 3
```

所有指针都指向同一个循环变量 `n`，而 `n` 在循环结束时是最后的值 `3`。

### 正确做法

**使用索引来获取元素的地址**（兼容所有版本）：

```go
nums := []int{1, 2, 3}
var pointers []*int

for i := range nums {
    pointers = append(pointers, &nums[i])
}

fmt.Println(*pointers[0], *pointers[1], *pointers[2]) // 输出: 1 2 3
```

### 版本说明

Go 1.22 起改进了循环变量的语义，每轮迭代的变量独立。因此在新版本中以下代码也是正确的：

```go
for _, n := range nums {
    pointers = append(pointers, &n) // Go 1.22+: 正确
}
```

但为了代码可移植性和养成良好习惯，**始终使用索引获取元素地址**是最佳实践。

---

## 陷阱三：映射遍历顺序随机

### 问题描述

映射的遍历顺序是**随机**的，不应依赖任何特定顺序。

### 正确做法

如果需要有序遍历，先对键进行排序：

```go
m := map[string]int{"b": 2, "a": 1, "c": 3}
keys := make([]string, 0, len(m))

for k := range m {
    keys = append(keys, k)
}
sort.Strings(keys)

for _, k := range keys {
    fmt.Println(k, m[k])
}
```

---

## 陷阱四：通道关闭后遍历

### 问题描述

当通道被关闭后，`range` 遍历会立即结束，收到所有已发送的值。

### 正确做法

确保只遍历需要的数据，或使用 `ok` 检查通道是否已关闭：

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
ch <- 3
close(ch)

for v := range ch {
    fmt.Println(v)
}
```

---

## 最佳实践总结

| 场景 | 建议做法 | 说明 |
|------|----------|------|
| 修改切片元素 | `for i := range nums { nums[i] = ... }` | 使用索引修改原元素 |
| 收集元素指针 | `for i := range nums { ptrs = append(ptrs, &nums[i]) }` | 使用索引获取稳定地址（Go 1.21 及更早版本必须，1.22+ 推荐） |
| 遍历映射有序输出 | 先对键排序，再遍历 | 映射遍历顺序随机 |
| 遍历通道 | 使用 `range ch` | 自动检测通道关闭 |

---

## 参考

- [Go 语言规范 - For 语句](https://go.dev/ref/spec#For_statements)
- [Go Blog - Go maps in action](https://go.dev/blog/maps)
