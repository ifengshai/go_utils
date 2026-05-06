package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"time"
)

func main() {

	////////////////////////////////////采集begin///////////////////////////////////////////////////
	// 创建一个新的上下文
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	// 设置浏览器选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		//设置chrome执行路径
		chromedp.ExecPath(`D:\workspace\tools\chrome-win\chrome.exe`),
		//设置user-agent
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36`),
		chromedp.Flag("headless", false),
		//禁用gpu
		chromedp.Flag("disable-gpu", false),
		//非自动化脚本执行
		chromedp.Flag("enable-automation", false),
		//是否使用扩展
		chromedp.Flag("disable-extensions", false),
		// 启用 JavaScript
		chromedp.Flag("enable-javascript", true),
		// 禁用 JavaScript
		//chromedp.Flag("disable-javascript", true),
		// 启用 JavaScript（默认）
		//// 设置代理服务器地址
		//chromedp.Flag("proxy-server", "your_proxy_server"),
		//// 禁用 HTTP 缓存
		//chromedp.Flag("disable-http-cache", true),
		//// 设置英语为首选语言
		//chromedp.Flag("lang", "en-US,en;q=0.9"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	// 创建一个浏览器实例
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()
	////设置上下文超时时间
	ctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	//存储html内容
	tempHtmlContent := ""

	//采集的网站
	href := "http://www.etmp3.com/"
	//执行爬虫任务
	err := chromedp.Run(
		ctx,
		chromedp.Navigate(href),                //导航到指定的网址（爬虫训练网站：https://scrape.center/）
		chromedp.WaitVisible(`input[id="wd"]`), //等待指定标签显示

		chromedp.SendKeys(`input[id="wd"]`, "my love"),   //设置关键词
		chromedp.Click("button.seh_b", chromedp.ByQuery), //点击搜索
		chromedp.Sleep(2), // 等待2秒，确保页面已更新
		//chromedp.WaitVisible("dev.pagedata"), //等待指定标签显示，同一个run事件中，等待标签都是等待第一个Navigate页面
		chromedp.InnerHTML("html", &tempHtmlContent),
	)
	if err != nil {
		fmt.Println("执行爬虫任务失败:", err)
	}

}
