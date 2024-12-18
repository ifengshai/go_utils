package main

import (
	"bufio"
	"fmt"
	"github.com/golang-module/carbon/v2"
	"os"
)

func main() {
	// 打开文件
	file, err := os.Open("./../file/aaa.txt")
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
		est := carbon.Parse(line, "PRC").SetTimezone("EST").ToDateTimeString()
		fmt.Println(est)
	}
	if scanner.Err() != nil {
		fmt.Println(scanner.Err())
	}
	fmt.Println("执行结束")
}
