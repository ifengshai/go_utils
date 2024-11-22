package main

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

func main() {
	// 创建一个上下文对象和取消函数
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	ptr := &ctx
	fmt.Printf("ctx内存地址：%p\n", ptr)
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
	ctxNew := context.WithValue(ctx, "s", "888")
	ctxNew.Value("s")
	fmt.Printf("%v\n", ctxNew.Value("s"))
	ptr := &ctxNew
	fmt.Printf("ctxNew内存地址：%p\n", ptr)

	fmt.Printf("类型判断：%v\n", reflect.TypeOf(reflect.TypeOf(reflect.TypeOf(ptr))))
	// 判断上下文是否已被取消
	select {
	case <-ctx.Done():
		// 上下文已被取消，说明超时
		fmt.Println("任务已超时，取消执行")
	default:
		// 模拟一个耗时的操作
		time.Sleep(4 * time.Second)
		// 执行正常完成
		fmt.Println("任务执行完毕")
	}
}
