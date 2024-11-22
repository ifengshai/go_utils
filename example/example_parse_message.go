package main

import (
	"bufio"
	"fmt"
	"github.com/tidwall/gjson"
	"os"
)

func main() {
	// 打开文件
	file, err := os.Open("./../file/message.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建bufio.Reader
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		//每行执行
		value := gjson.Get(line, "create_time")
		fmt.Println(value)

	}
	if scanner.Err() != nil {
		fmt.Println(scanner.Err())
	}
	fmt.Println("执行结束")
}
