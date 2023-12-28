package main

import (
	"fmt"
	"github.com/ifengshai/go_utils"
	"net/url"
	"time"
)

func main() {
	urlLink := "http://www.etmp3.com/js/play.php"
	method := "POST"
	headers := map[string]string{
		"content-Type": "application/x-www-form-urlencoded",
		"accept":       "application/json, text/plain, */*",
		"origin":       "http://www.etmp3.com",
		"referer":      "http://www.etmp3.com/mp3/216053830.html",
	}
	body := map[string]string{
		"id":   "216053830",
		"type": "mp3",
	}
	timeout := 3 * time.Second
	retryCount := 2
	retryInterval := 1 * time.Second

	values := url.Values{}
	for k, v := range body {
		values.Add(k, v)
	}

	response, err := go_utils.SendHTTPRequest(urlLink, method, headers, []byte(values.Encode()),
		timeout, retryCount, retryInterval, true)
	if err != nil {
		fmt.Printf("请求出错：%s\n", err.Error())
		return
	}

	fmt.Printf("响应内容：%s\n", string(response))

}
