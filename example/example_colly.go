package main

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
	"time"
)

func main() {

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
	ctx, cancel = context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	colly(ctx, "https://spa2.scrape.center/")

}

func colly(ctx context.Context, href string) {
	fmt.Println("################################################################################")
	//存储html内容
	htmlContent := ""
	// 创建一个空的 Node 切片，用于存储 li 元素
	var liNodes []*cdp.Node
	//执行爬虫任务
	err := chromedp.Run(
		ctx,
		chromedp.Navigate(href), //导航到指定的网址（爬虫训练网站：https://scrape.center/）
		chromedp.WaitVisible("ul.el-pager li.number", chromedp.ByQuery), //等待指定标签显示
		chromedp.InnerHTML("html", &htmlContent),

		// 获取所有页码
		chromedp.Nodes("ul.el-pager li.number", &liNodes, chromedp.ByQueryAll),

		chromedp.ActionFunc(func(ctx context.Context) error {
			var tempText string
			var tempKey int
			for key, liNode := range liNodes {
				if strings.Contains(liNode.AttributeValue("class"), "active") {
					tempKey = key
				}
			}
			if tempKey+1 < len(liNodes) {

				_ = chromedp.Run(ctx,
					chromedp.Text(liNodes[tempKey+1].FullXPath(), &tempText, chromedp.BySearch),
				)

				//下一页内容
				var tempHtmlContent string
				//下一页html
				var tempHref string

				_ = chromedp.Run(
					ctx,
					chromedp.Click(liNodes[tempKey+1].FullXPath()),                  //点击下一页
					chromedp.WaitVisible(".el-pagination__total", chromedp.ByQuery), //等待指定标签显示
					chromedp.InnerHTML("html", &tempHtmlContent),                    //当前页内容
					chromedp.Location(&tempHref),                                    //下一页html
				)
				fmt.Println("当前页内容：")
				getText(tempHtmlContent)
				fmt.Printf("下一页是第%s页\n", tempText)
				fmt.Printf("下一页url:%s\n", tempHref)
				colly(ctx, tempHref)

			}
			return nil
		}),
	)
	if err != nil {
		fmt.Println("执行爬虫任务失败:", err)
	}

}

func getText(str string) {
	// 将响应的Body传递给goquery.LoadDocument函数进行解析
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(str))
	if err != nil {
		fmt.Println("读取html内容失败:", err)
		log.Fatal(err)
	}
	// 使用CSS选择器来查找元素并遍历结果
	doc.Find("h2.m-b-sm").Each(func(i int, s *goquery.Selection) {
		fmt.Println("内容:", s.Text())
	})
	return
}
