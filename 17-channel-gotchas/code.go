package main

import (
	"fmt"
	"time"
)

// 演示 select 随机性
func demonstrateSelectRandomness() {
	fmt.Println("\n=== select 随机性 ===")

	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)

	ch1Count, ch2Count := 0, 0

	// 循环多次，观察 select 的随机选择
	for i := 0; i < 20; i++ {
		// 重置 channel 状态
		select {
		case <-ch1:
		case <-ch2:
		default:
		}

		// 向两个 channel 发送数据
		select {
		case ch1 <- 1:
		case ch2 <- 2:
		}

		// 从可用的 channel 接收
		select {
		case <-ch1:
			ch1Count++
		case <-ch2:
			ch2Count++
		}
	}

	fmt.Printf("统计结果: ch1 被选中 %d 次, ch2 被选中 %d 次\n", ch1Count, ch2Count)
	fmt.Println("可以看到两个 channel 都有被选中，证明 select 是随机选择的")
}

// 演示 nil channel 永久阻塞（注释说明，不实际运行阻塞代码）
func demonstrateNilChannelBlocking() {
	fmt.Println("\n=== nil channel 永久阻塞 ===")
	fmt.Println("nil channel 的发送和接收操作会永久阻塞")
	fmt.Println("示例代码（已注释以避免永久阻塞）：")

	fmt.Println(`
	var ch chan int // ch 是 nil channel

	// 下面的操作会永久阻塞：
	// <-ch  // 接收操作永久阻塞
	// ch <- 1 // 发送操作永久阻塞

	// nil channel 的特性可以用于 select 语句中
	// 当 case 中的 channel 为 nil 时，该 case 会被忽略
	var nilCh chan int
	select {
	case v := <-nilCh: // 这个 case 永远不会被选中
		fmt.Println(v)
	case v := <-make(chan int):
		fmt.Println("这个会执行")
	}
	`)
}

// 演示 closed channel 仍可接收数据
func demonstrateClosedChannelReceiving() {
	fmt.Println("\n=== closed channel 仍可接收数据 ===")

	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	ch <- 30
	close(ch) // 关闭 channel

	fmt.Println("从已关闭的 channel 接收数据：")
	for v := range ch {
		fmt.Printf("接收: %d\n", v)
	}

	// 继续尝试接收（返回零值）
	fmt.Println("\nchannel 关闭后继续接收：")
	v, ok := <-ch
	fmt.Printf("值: %d, ok: %v (ok=false 表示 channel 已关闭)\n", v, ok)
}

// 演示使用 ok 判断 channel 状态
func demonstrateOkStatusCheck() {
	fmt.Println("\n=== 使用 ok 判断 channel 状态 ===")

	// 情况1：正常接收
	ch1 := make(chan int, 1)
	ch1 <- 100
	close(ch1)

	v, ok := <-ch1
	fmt.Printf("正常接收 - 值: %d, ok: %v\n", v, ok)

	v, ok = <-ch1
	fmt.Printf("channel 已关闭 - 值: %d, ok: %v\n", v, ok)

	// 情况2：nil channel（在 select 中不会阻塞，会被忽略）
	var nilCh chan int
	fmt.Println("\nnil channel 说明：")
	fmt.Println("  - 直接接收 <-nilCh 会永久阻塞")
	fmt.Println("  - 在 select 中的 nil channel case 会被忽略")

	// 使用 default case 避免阻塞
	select {
	case v, ok := <-nilCh:
		fmt.Printf("nil channel - 值: %d, ok: %v\n", v, ok)
	default:
		fmt.Println("nil case 被忽略，default 分支执行")
	}

	// 情况3：通过 select 使用 ok
	fmt.Println("\n通过 select 判断 channel 状态：")
	ch2 := make(chan int, 1)
	ch2 <- 200
	close(ch2)

	for {
		select {
		case v, ok := <-ch2:
			if !ok {
				fmt.Println("channel 已关闭，退出循环")
				return
			}
			fmt.Printf("接收到值: %d\n", v)
		}
	}
}

// 演示向已关闭 channel 发送数据会 panic
func demonstrateSendOnClosedChannel() {
	fmt.Println("\n=== 向已关闭 channel 发送数据会 panic ===")
	fmt.Println("下面的代码会 panic，请勿在实际代码中这样做：")

	fmt.Println(`
	ch := make(chan int)
	close(ch)
	ch <- 1 // panic: send on closed channel
	`)
}

// 演示重复关闭 channel 会 panic
func demonstrateDoubleCloseChannel() {
	fmt.Println("\n=== 重复关闭 channel 会 panic ===")
	fmt.Println("下面的代码会 panic，请勿在实际代码中这样做：")

	fmt.Println(`
	ch := make(chan int)
	close(ch)
	close(ch) // panic: close of closed channel
	`)
}

// 演示带超时的 select 防止永久阻塞
func demonstrateSelectWithTimeout() {
	fmt.Println("\n=== select 超时机制 ===")

	var nilCh chan int // nil channel

	select {
	case v := <-nilCh:
		fmt.Printf("从 nil channel 接收: %d\n", v)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("操作超时（500ms）")
	}

	// 正常 channel 带超时
	ch := make(chan int, 1)
	select {
	case v := <-ch:
		fmt.Printf("从 ch 接收: %d\n", v)
	case <-time.After(500 * time.Millisecond):
		fmt.Println("从 ch 接收超时")
	}
}

func main() {
	fmt.Println("========================================")
	fmt.Println("     Go Channel 注意事项演示")
	fmt.Println("========================================")

	demonstrateSelectRandomness()
	demonstrateNilChannelBlocking()
	demonstrateClosedChannelReceiving()
	demonstrateOkStatusCheck()
	demonstrateSendOnClosedChannel()
	demonstrateDoubleCloseChannel()
	demonstrateSelectWithTimeout()

	fmt.Println("\n========================================")
	fmt.Println("     演示完成")
	fmt.Println("========================================")
}