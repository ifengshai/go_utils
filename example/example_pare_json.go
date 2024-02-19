package main

import (
	"bufio"
	"fmt"
	"github.com/tidwall/gjson"
	"os"
)

func main() {

	file, err := os.Open("./../file/a.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		//每行执行
		value := gjson.Get(line, "data.app_index_product.products.#(group_id == \"99999999\")#.products.#.id")
		fmt.Println(value)
		os.Exit(1)
	}
	if scanner.Err() != nil {
		fmt.Println(scanner.Err())
	}
	fmt.Println("执行结束")

}
