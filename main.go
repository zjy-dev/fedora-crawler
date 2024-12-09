package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
	"github.com/zjy-dev/fedora-crawler/internal/crawler"
)

var startPrefix string = "0a"
var endPrefix string = "zz"

var x64RpmUrlFilePath = ""
var aarch64RpmUrlFilePath = ""

func main() {
	if len(os.Args) != 3 {
		fmt.Println("用法: ./xxx path/to/fedora_rpm_url_x64.txt path/to/fedora_rpm_url_aarch64.txt")
		return
	}
	x64RpmUrlFilePath = os.Args[1]
	aarch64RpmUrlFilePath = os.Args[2]
	c := colly.NewCollector(
		colly.AllowedDomains("packages.fedoraproject.org", "koji.fedoraproject.org"),
		colly.MaxDepth(1),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",             // 适用于所有域名
		Delay:       1 * time.Second, // 设置 1 秒的间隔
		RandomDelay: 0,               // 不使用随机延迟
	})
	color.Cyan("开始爬取包前缀\n")
	_, prefixHrefs := crawlPrefixs(c)

	color.Cyan("开始爬取 RPM 包\n")
	crawlPackages(c, prefixHrefs)
}

func crawlPackages(c *colly.Collector, prefixHrefs []string) {
	mainPageUrlsCrawler := crawler.MainPageUrlsCrawler(c.Clone())
	mainPageCrawler := crawler.MainPageCrawler(c.Clone())
	buildsPageCrawler := crawler.BuildsPageCrawler(c.Clone())
	downloadPageCrawler := crawler.DownloadPageCrawler(c.Clone())
	for i := 0; i < len(prefixHrefs); i++ {
		// 爬取软件包主页 URLs
		packageNames, mainPageUrls := mainPageUrlsCrawler(prefixHrefs[i])
		color.Blue("爬取软件包主页 URLs, 包名列表: %v, 主页 URL 列表: %v", packageNames, mainPageUrls)

		for i := 0; i < len(mainPageUrls); i++ {
			// 爬取软件包 Builds 页
			buildsUrl := mainPageCrawler(mainPageUrls[i])
			// color.Blue("爬取软件包 Builds 页: %v", buildsUrl)

			// 爬取 Builds 页最新下载页地址
			latestDownloadPageURL := buildsPageCrawler(buildsUrl)
			// color.Blue("爬取 Builds 页最新下载页地址: %v", latestDownloadPageURL)

			x64RpmURL, aarch64RpmURL := downloadPageCrawler(latestDownloadPageURL)
			if len(x64RpmURL) != 0 {
				color.Yellow("爬取 x64 rpm 包 %v 下载地址: %v", packageNames[i], x64RpmURL)
				appendToFile(x64RpmURL, x64RpmUrlFilePath)
			}
			if len(aarch64RpmURL) != 0 {
				color.Yellow("爬取 aarch64 rpm 包 %v 下载地址: %v", packageNames[i], aarch64RpmURL)
				appendToFile(aarch64RpmURL, aarch64RpmUrlFilePath)
			}
		}
	}
}

func appendToFile(content string, filepath string) {
	// 打开文件（如果文件不存在则创建，使用 os.O_APPEND 来追加）
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("打开文件失败:", err)
		return
	}
	defer file.Close()

	// 写入内容到文件
	_, err = file.WriteString(content + "\n")
	if err != nil {
		fmt.Println("写入文件失败:", err)
		return
	}
}

func crawlPrefixs(c *colly.Collector) (prefixs []string, prefixHrefs []string) {
	c.OnHTML("#prefix", func(e *colly.HTMLElement) {
		e.ForEach("a", func(i int, e *colly.HTMLElement) {
			prefixs = append(prefixs, e.Text)
			prefixHrefs = append(prefixHrefs, e.Request.AbsoluteURL(e.Attr("href")))
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		color.Red("访问出错, url: %s, err: %v\n", r.Request.URL, err)
	})

	c.Visit("https://packages.fedoraproject.org/index-static.html")

	startIdx, endIdx := -1, -1
	for i := 0; i < len(prefixs); i++ {
		if prefixs[i] == startPrefix {
			startIdx = i
		}
		if prefixs[i] == endPrefix {
			endIdx = i
		}
	}
	if startIdx == -1 || endIdx == -1 {
		log.Panicf("startPrefix: %v or endPrefix: %v not found\n", startPrefix, endPrefix)
	}
	prefixs = prefixs[startIdx : endIdx+1]
	prefixHrefs = prefixHrefs[startIdx : endIdx+1]
	return
}
