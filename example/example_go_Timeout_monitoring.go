package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// 创建一个上下文对象和取消函数
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 启动一个 goroutine 执行任务
	go doTask(ctx)

	// 等待任务完成或超时
	select {
	case <-ctx.Done():
		// 执行超时
		fmt.Println("任务执行超时")
	case <-time.After(5 * time.Second):
		// 超过等待时间，认为任务完成
		fmt.Println("任务执行完成")
	}
}

func doTask(ctx context.Context) {
	// 模拟一个耗时的操作
	time.Sleep(4 * time.Second)

	// 判断上下文是否已被取消
	select {
	case <-ctx.Done():
		// 上下文已被取消，说明超时
		fmt.Println("任务已超时，取消执行")
	default:
		// 执行正常完成
		fmt.Println("任务执行完毕")
	}
}
