package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	// 打开文件
	file, err := os.Open("./../file/temp.json")
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return
	}
	defer file.Close()

	// 创建一个 Scanner 对象来读取文件
	scanner := bufio.NewScanner(file)

	// 定义一个结构体来存储 JSON 数据
	type JsonStruct struct {
		OrderId       string `json:"OrderId"`
		CurrencyCode  string `json:"CurrencyCode"`
		ItemSubTotal1 string `json:"ItemSubTotal1"`
		EventDate     string `json:"EventDate"`
	}
	type OptionStruct struct {
		Json JsonStruct `json:"json"`
	}
	type ContextStruct struct {
		OptionStruct OptionStruct `json:"options"`
	}
	type MyStruct struct {
		TraceId string        `json:"trace_id"`
		Context ContextStruct `json:"context"`
	}
	var myStruct = MyStruct{}

	// 创建一个文件并打开以写入 CSV 数据
	csvFile, err := os.Create("./../file/output.csv")
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	// 创建一个 CSV writer
	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// 逐行读取文件内容
	for scanner.Scan() {
		var stringSlice = make([]string, 0, 0)
		line := scanner.Text()
		err := json.Unmarshal([]byte(line), &myStruct)
		if err != nil {
			fmt.Println("转换为 JSON 出错:", err)
			return
		}

		stringSlice = append(stringSlice, myStruct.Context.OptionStruct.Json.OrderId)
		stringSlice = append(stringSlice, myStruct.Context.OptionStruct.Json.CurrencyCode)
		stringSlice = append(stringSlice, myStruct.Context.OptionStruct.Json.ItemSubTotal1)
		stringSlice = append(stringSlice, myStruct.Context.OptionStruct.Json.EventDate)

		writer.Write(stringSlice)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("文件读取错误:", err)
	}
}
