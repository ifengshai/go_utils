package main

import (
	"fmt"
	"time"
)

func main() {
	// 定义需要转换的北京时间字符串列表
	timeStrings := []string{
		"2025-11-11 14:37:42",
		"2025-11-12 09:15:03",
		"2025-11-13 18:54:27",
		"2025-11-14 03:22:58",
		"2025-11-15 11:48:19",
		"2025-11-16 06:05:44",
		"2025-11-17 16:29:11",
		"2025-11-18 01:57:35",
		"2025-11-19 19:03:06",
		"2025-11-20 07:41:53",
		"2025-11-21 12:16:28",
		"2025-11-22 05:38:09",
		"2025-11-23 15:09:51",
		"2025-11-24 10:27:14",
		"2025-11-25 20:00:00",
		"2025-11-26 08:33:47",
		"2025-11-27 13:48:25",
		"2025-11-28 04:11:32",
	}

	// 定义美东时区（UTC-5）
	beijingLocation, err := time.LoadLocation("America/New_York")
	if err != nil {
		fmt.Printf("无法加载时区: %v\n", err)
		return
	}

	// 循环转换每个时间字符串为时间戳
	fmt.Println("北京时间 -> 时间戳(秒) -> 时间戳(毫秒)")
	fmt.Println("---------------------------------------")
	for _, ts := range timeStrings {
		// 解析时间字符串（使用北京时间时区）
		t, err := time.ParseInLocation("2006-01-02 15:04:05", ts, beijingLocation)
		if err != nil {
			fmt.Printf("解析时间 %s 失败: %v\n", ts, err)
			continue
		}

		// 转换为时间戳（秒和毫秒）
		timestampSec := t.Unix()
		timestampMs := t.UnixMilli()

		// 输出结果
		fmt.Printf("%s -> %d -> %d\n", ts, timestampSec, timestampMs)
	}
}
