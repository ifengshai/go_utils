package main

import (
	"fmt"
	"github.com/ifengshai/go_utils"
	"net/url"
	"time"
)

func main() {
	urlLink := "https://api.44h4.com/lc.php?cid=216053830"
	method := "POST"
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	body := map[string]string{
		"name": "Alice",
		"age":  "25",
	}
	timeout := 3 * time.Second
	retryCount := 2
	retryInterval := 1 * time.Second

	values := url.Values{}
	for k, v := range body {
		values.Set(k, v)
	}

	response, err := go_utils.SendHTTPRequest(urlLink, method, headers, []byte(values.Encode()),
		timeout, retryCount, retryInterval, true)
	if err != nil {
		fmt.Printf("请求出错：%s\n", err.Error())
		return
	}

	fmt.Printf("响应内容：%s\n", string(response))
}
