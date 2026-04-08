package main

import (
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// 演示 1: Goroutine 创建先于执行
// =============================================================================

func demo1_goroutineCreationPrecedesExecution() {
	fmt.Println("=== Goroutine 创建先于执行 ===")

	var msg string

	go func() {
		// 由于 goroutine 创建 happens-before 其开始执行
		// 此处看到的 msg 可能是空字符串（如果创建在赋值之前）
		// 也可能是 "hello"（如果调度导致赋值先执行）
		fmt.Printf("goroutine sees: %q\n", msg)
	}()

	msg = "hello"
	time.Sleep(10 * time.Millisecond)
	fmt.Println()
}

// =============================================================================
// 演示 2: Goroutine 退出无保证
// =============================================================================

func demo2_goroutineExitNotGuaranteed() {
	fmt.Println("=== Goroutine 退出无保证 ===")

	go func() {
		time.Sleep(50 * time.Millisecond)
		fmt.Println("goroutine: done sleeping")
	}()

	fmt.Println("main: about to exit (goroutine may not finish)")
	// 主函数退出时，goroutine 可能还没打印
	fmt.Println()
}

// =============================================================================
// 演示 3: Channel Send 先于 Receive
// =============================================================================

func demo3_channelSendBeforeReceive() {
	fmt.Println("=== Channel Send 先于 Receive ===")

	ch := make(chan int)

	go func() {
		fmt.Println("sender: about to send 42")
		ch <- 42
		fmt.Println("sender: sent 42")
	}()

	go func() {
		val := <-ch
		fmt.Printf("receiver: got %d\n", val)
	}()

	time.Sleep(20 * time.Millisecond)
	fmt.Println()
}

// =============================================================================
// 演示 4: Channel Close 先于 Receive
// =============================================================================

func demo4_channelCloseBeforeReceive() {
	fmt.Println("=== Channel Close 先于 Receive ===")

	ch := make(chan int)

	go func() {
		fmt.Println("closer: closing channel")
		close(ch)
		fmt.Println("closer: closed")
	}()

	go func() {
		val, ok := <-ch
		// channel 关闭后，接收会得到零值和 false
		fmt.Printf("receiver: val=%d, ok=%v (channel closed)\n", val, ok)
	}()

	time.Sleep(20 * time.Millisecond)
	fmt.Println()
}

// =============================================================================
// 演示 5: Unbuffered Channel: Receive 先于 Send
// =============================================================================

func demo5_unbufferedReceiveBeforeSend() {
	fmt.Println("=== Unbuffered Channel: Receive 先于 Send ===")

	ch := make(chan int)

	go func() {
		val := <-ch
		fmt.Printf("receiver: got %d\n", val)
	}()

	go func() {
		// Send 会阻塞直到有人接收
		// 这证明了 unbuffered channel: receive 先于 send 完成
		fmt.Println("sender: about to send 100")
		ch <- 100
		fmt.Println("sender: send completed")
	}()

	time.Sleep(20 * time.Millisecond)
	fmt.Println()
}

// =============================================================================
// 演示 6: 使用 WaitGroup 正确等待 Goroutine
// =============================================================================

func demo6_waitgroupForWaiting() {
	fmt.Println("=== 使用 WaitGroup 等待 Goroutine ===")

	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			fmt.Printf("goroutine %d done\n", id)
		}(i)
	}

	fmt.Println("main: waiting for goroutines...")
	wg.Wait() // 正确等待所有 goroutine 完成
	fmt.Println("main: all goroutines finished")
	fmt.Println()
}

// =============================================================================
// 演示 7: Buffered Channel 不保证 Send/Receive 顺序
// =============================================================================

func demo7_bufferedChannelNoGuarantee() {
	fmt.Println("=== Buffered Channel 无此保证 ===")

	ch := make(chan int, 1)

	ch <- 1 // 不阻塞，因为有缓冲

	val := <-ch
	fmt.Printf("buffered channel: got %d\n", val)
	fmt.Println("( buffered channel 不提供 receive 先于 send 的保证 )")
	fmt.Println()
}

// =============================================================================
// 演示 8: 演示 Happens-Before 的实际应用
// =============================================================================

func demo8_happensBeforeApplication() {
	fmt.Println("=== Happens-Before 应用：正确的数据同步 ===")

	ch := make(chan int)

	// 正确模式：使用 channel 同步
	go func() {
		for i := 0; i < 5; i++ {
			ch <- i
		}
		close(ch)
	}()

	for v := range ch {
		fmt.Printf("main: received %d\n", v)
	}

	fmt.Println()
}

func main() {
	demo1_goroutineCreationPrecedesExecution()
	demo2_goroutineExitNotGuaranteed()
	demo3_channelSendBeforeReceive()
	demo4_channelCloseBeforeReceive()
	demo5_unbufferedReceiveBeforeSend()
	demo6_waitgroupForWaiting()
	demo7_bufferedChannelNoGuarantee()
	demo8_happensBeforeApplication()

	fmt.Println("=== 所有演示完成 ===")
}
