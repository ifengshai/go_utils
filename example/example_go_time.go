package main

import (
	"github.com/golang-module/carbon/v2"
	"github.com/ifengshai/go_utils"
)

func main() {

	time := carbon.Now().SubDays(330).ToDateString()
	go_utils.Printf(time)
	timeSliceString := []string{
		carbon.Now().ToString(),         // 2020-08-05 13:14:15 +0800 CST
		carbon.Now().ToDateTimeString(), // 2020-08-05 13:14:15
		// 今天日期
		carbon.Now().ToDateString(), // 2020-08-05
		// 今天时间
		carbon.Now().ToTimeString(), // 13:14:15
		// 指定时区的今天此刻
		carbon.Now(carbon.NewYork).ToDateTimeString(), // 2020-08-05 14:14:15
	}
	go_utils.Printf(timeSliceString)

	timeSliceInt := []int64{
		// 今天毫秒级时间戳
		carbon.Now().TimestampMilli(), // 1596604455999
		// 今天微秒级时间戳
		carbon.Now().TimestampMicro(), // 1596604455999999
		// 今天纳秒级时间戳
		carbon.Now().TimestampNano(), // 1596604455999999999
	}
	go_utils.Printf(timeSliceInt)

}
