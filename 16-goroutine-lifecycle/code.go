// 代码示例：Goroutine 生命周期管理
// 运行方式：go run code.go
//
// 本文件包含以下示例：
// 1. Goroutine 泄漏问题演示
// 2. Context 取消模式
// 3. Channel 关闭信号模式
// 4. Worker Pool 模式
// 5. WaitGroup 等待模式

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// 示例1：Goroutine 泄漏问题演示
// ============================================================================

// 内存泄漏的 goroutine - 永远等待 channel
func leakyGoroutine() {
	ch := make(chan int)
	go func() {
		val := <-ch // 如果没有人发送数据，这里会永久阻塞
		fmt.Println("Received:", val)
	}()
	// ch 从来没有被写入，goroutine 永远阻塞在 <-ch
	// 这就是 goroutine 泄漏 - 它将永远存在于内存中
}

// 修复版本：使用 select 和 default 分支
func nonLeakyGoroutine() {
	ch := make(chan int)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case val := <-ch:
				fmt.Println("Received:", val)
			case <-done:
				fmt.Println("Goroutine 退出")
				return
			default:
				// 非阻塞，继续做其他工作
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// 模拟发送数据
	ch <- 42
	close(done) // 发送退出信号
}

// ============================================================================
// 示例2：Context 取消模式（推荐）
// ============================================================================

// fetchData 模拟数据获取，使用 context 进行取消
func fetchData(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			// context 被取消时执行
			return ctx.Err()
		default:
			// 模拟工作
			time.Sleep(200 * time.Millisecond)
			fmt.Println("正在获取数据...")
		}
	}
}

// contextWithTimeout 演示带超时的 context
func contextWithTimeout() {
	fmt.Println("\n--- 演示：Context 超时取消 ---")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := fetchData(ctx)
	if err != nil {
		fmt.Printf("操作被取消: %v\n", err)
	}
}

// contextWithCancel 演示手动取消
func contextWithCancel() {
	fmt.Println("\n--- 演示：Context 手动取消 ---")

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel() // 手动触发取消
	}()

	err := fetchData(ctx)
	if err != nil {
		fmt.Printf("操作被取消: %v\n", err)
	}
}

// ============================================================================
// 示例3：Channel 关闭信号模式
// ============================================================================

// workerWithChannel 演示使用 channel 信号通知退出
func workerWithChannel(id int, done chan struct{}, jobs <-chan int, results chan<- int) {
	defer fmt.Printf("Worker %d: 退出\n", id)
	for {
		select {
		case <-done:
			// 收到退出信号
			return
		case job, ok := <-jobs:
			if !ok {
				// jobs channel 已关闭
				return
			}
			// 处理任务
			time.Sleep(100 * time.Millisecond)
			results <- job * 2
		}
	}
}

// channelClosePattern 演示 channel 关闭模式
func channelClosePattern() {
	fmt.Println("\n--- 演示：Channel 关闭信号模式 ---")

	jobs := make(chan int, 10)
	results := make(chan int, 10)
	done := make(chan struct{})

	// 启动 3 个 worker
	for i := 0; i < 3; i++ {
		go workerWithChannel(i, done, jobs, results)
	}

	// 发送任务
	for i := 1; i <= 5; i++ {
		jobs <- i
	}

	// 关闭 jobs channel，通知 worker 没有更多任务
	close(jobs)

	// 等待 results
	go func() {
		for result := range results {
			fmt.Printf("结果: %d\n", result)
		}
	}()

	// 等待所有结果处理完成
	time.Sleep(1 * time.Second)

	// 发送退出信号给所有 worker
	close(done)

	// 等待 worker 退出
	time.Sleep(500 * time.Millisecond)
}

// ============================================================================
// 示例4：Worker Pool 模式
// ============================================================================

// Job 代表一个工作单元
type Job struct {
	ID   int
	Data string
}

// Result 代表工作结果
type Result struct {
	JobID  int
	Output string
}

// workerPoolWithContext 演示带 context 的 worker pool
func workerPoolWithContext(ctx context.Context, workerID int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	defer fmt.Printf("Worker %d: 退出\n", workerID)

	for {
		select {
		case <-ctx.Done():
			// 收到取消信号
			return
		case job, ok := <-jobs:
			if !ok {
				// jobs channel 已关闭
				return
			}
			// 模拟处理
			time.Sleep(100 * time.Millisecond)
			results <- Result{
				JobID:  job.ID,
				Output: fmt.Sprintf("处理: %s", job.Data),
			}
		}
	}
}

// workerPoolPattern 演示 worker pool 模式
func workerPoolPattern() {
	fmt.Println("\n--- 演示：Worker Pool 模式 ---")

	const numWorkers = 3
	const numJobs = 10

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	var wg sync.WaitGroup

	// 启动 worker pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go workerPoolWithContext(ctx, i, jobs, results, &wg)
	}

	// 发送任务
	go func() {
		for i := 0; i < numJobs; i++ {
			jobs <- Job{ID: i, Data: fmt.Sprintf("任务-%d", i)}
		}
		close(jobs)
	}()

	// 等待结果
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	resultCount := 0
	for result := range results {
		fmt.Printf("收到结果: JobID=%d, Output=%s\n", result.JobID, result.Output)
		resultCount++
	}

	fmt.Printf("共收到 %d 个结果\n", resultCount)

	select {
	case <-ctx.Done():
		fmt.Println("因超时取消")
	default:
		fmt.Println("所有任务完成")
	}
}

// ============================================================================
// 示例5：WaitGroup 等待模式
// ============================================================================

// longRunningTask 模拟一个长时间运行的任务
func longRunningTask(id int, duration time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	defer fmt.Printf("Task %d: 完成\n", id)

	fmt.Printf("Task %d: 开始，耗时 %v\n", id, duration)
	time.Sleep(duration)
}

// waitGroupPattern 演示 WaitGroup 等待模式
func waitGroupPattern() {
	fmt.Println("\n--- 演示：WaitGroup 等待模式 ---")

	var wg sync.WaitGroup

	// 启动多个任务
	tasks := []struct {
		id       int
		duration time.Duration
	}{
		{1, 500 * time.Millisecond},
		{2, 300 * time.Millisecond},
		{3, 800 * time.Millisecond},
	}

	for _, task := range tasks {
		wg.Add(1)
		go longRunningTask(task.id, task.duration, &wg)
	}

	// 等待所有任务完成
	fmt.Println("等待所有任务完成...")
	wg.Wait()
	fmt.Println("所有任务已完成！")
}

// ============================================================================
// 示例6：多个 goroutine 共享同一个 Context
// ============================================================================

// sharedContextWorker 共享 context 的 worker
func sharedContextWorker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d: 收到取消信号，原因: %v\n", id, ctx.Err())
			return
		case <-time.After(300 * time.Millisecond):
			fmt.Printf("Worker %d: 执行中...\n", id)
		}
	}
}

// sharedContextPattern 演示多个 goroutine 共享同一个 context
func sharedContextPattern() {
	fmt.Println("\n--- 演示：多个 goroutine 共享 Context ---")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// 启动 3 个共享 context 的 worker
	for i := 0; i < 3; i++ {
		go sharedContextWorker(ctx, i)
	}

	// 等待 context 超时
	<-ctx.Done()
	time.Sleep(100 * time.Millisecond) // 等待日志输出
}

// ============================================================================
// 示例7：正确的资源清理模式
// ============================================================================

// ResourceCleaner 演示正确的资源清理
type ResourceCleaner struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewResourceCleaner() *ResourceCleaner {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return &ResourceCleaner{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (rc *ResourceCleaner) Start() {
	fmt.Println("\n--- 演示：正确资源清理模式 ---")

	// 启动后台任务
	rc.wg.Add(1)
	go func() {
		defer rc.wg.Done()
		for {
			select {
			case <-rc.ctx.Done():
				fmt.Println("后台任务: 收到取消信号，退出")
				return
			default:
				time.Sleep(200 * time.Millisecond)
				fmt.Println("后台任务: 运行中...")
			}
		}
	}()
}

func (rc *ResourceCleaner) Stop() {
	fmt.Println("执行清理...")
	rc.cancel()      // 取消 context
	rc.wg.Wait()     // 等待所有 goroutine 退出
	fmt.Println("清理完成")
}

func resourceCleanupPattern() {
	cleaner := NewResourceCleaner()
	cleaner.Start()

	// 模拟工作
	time.Sleep(1 * time.Second)

	// 主动停止
	cleaner.Stop()
}

// ============================================================================
// 主函数演示
// ============================================================================

func main() {
	fmt.Println("========================================")
	fmt.Println("Goroutine 生命周期管理 - 代码演示")
	fmt.Println("========================================")

	// 演示1：Context 超时取消
	contextWithTimeout()

	// 演示2：Context 手动取消
	contextWithCancel()

	// 演示3：Channel 关闭信号模式
	channelClosePattern()

	// 演示4：Worker Pool 模式
	workerPoolPattern()

	// 演示5：WaitGroup 等待模式
	waitGroupPattern()

	// 演示6：多个 goroutine 共享 Context
	sharedContextPattern()

	// 演示7：正确的资源清理模式
	resourceCleanupPattern()

	fmt.Println("\n========================================")
	fmt.Println("总结：Goroutine 生命周期管理要点")
	fmt.Println("========================================")
	fmt.Println("1. 使用 context.Context 进行取消操作（推荐）")
	fmt.Println("2. 使用 sync.WaitGroup 等待多个 goroutine 完成")
	fmt.Println("3. 使用 worker pool 限制并发数量")
	fmt.Println("4. 使用 channel 关闭作为退出信号")
	fmt.Println("5. 确保所有 goroutine 都有明确的退出机制")
	fmt.Println("6. 避免 goroutine 泄漏：永远不让 goroutine 无限阻塞")
}
