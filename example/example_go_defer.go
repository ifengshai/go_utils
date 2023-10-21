package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("return:", test()) // return 后执行defer defer执行完成外部调用才能拿到返回值
}

func test() (i int) { //这里返回值没有命名

	defer func() {
		time.Sleep(time.Second * 5)
		i++
		fmt.Println("defer1", i) //作为闭包引用的话，则会在defer函数执行时根据整个上下文确定当前的值。i=2
	}()
	defer func() {
		i++
		fmt.Println("defer2", i) //作为闭包引用的话，则会在defer函数执行时根据整个上下文确定当前的值。i=1
	}()
	i = 1
	return 5
}
