package crawler

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/gocolly/colly/v2"
)

func MainPageUrlsCrawler(c *colly.Collector) func(string) ([]string, []string) {
	var packageNames, mainPageUrls []string

	c.OnHTML("body > div > ul", func(e *colly.HTMLElement) {
		e.ForEach("li > a", func(i int, e *colly.HTMLElement) {
			packageNames = append(packageNames, e.Text)
			mainPageUrls = append(mainPageUrls, e.Request.AbsoluteURL(e.Attr("href")))
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		color.Red("访问出错, url: %s, err: %v\n", r.Request.URL, err)
	})
	return func(url string) ([]string, []string) {
		packageNames, mainPageUrls = make([]string, 0), make([]string, 0)
		c.Visit(url)
		return packageNames, mainPageUrls
	}
}

func MainPageCrawler(c *colly.Collector) func(string) string {
	buildsUrl := ""
	c.OnHTML("body > div > ul > li:nth-child(1) > a", func(e *colly.HTMLElement) {
		if strings.TrimSpace(e.Text) != "Builds" {
			log.Panicln("主页爬取的选择器有问题, 爬到的不是 Builds 按钮")
		}
		buildsUrl = e.Request.AbsoluteURL(e.Attr("href"))
	})

	c.OnError(func(r *colly.Response, err error) {
		color.Red("访问出错, url: %s, err: %v\n", r.Request.URL, err)
	})

	return func(url string) string {
		c.Visit(url)
		return buildsUrl
	}
}

func BuildsPageCrawler(c *colly.Collector) func(string) string {
	latestVersionDownloadPageURL := ""
	c.OnHTML("#packages > div:nth-child(2) > div.container.main.pt-3 > table > tbody > tr:nth-child(3) > td > table > tbody > tr:nth-child(3) > td:nth-child(1) > a", func(e *colly.HTMLElement) {
		latestVersionDownloadPageURL = e.Request.AbsoluteURL(e.Attr("href"))
	})

	c.OnError(func(r *colly.Response, err error) {
		color.Red("访问出错, url: %s, err: %v\n", r.Request.URL, err)
	})

	return func(url string) string {
		c.Visit(url)
		return latestVersionDownloadPageURL
	}
}

func DownloadPageCrawler(c *colly.Collector) func(string) (string, string) {
	x64RpmDownloadURL, aarch64RpmDownloadURL := "", ""

	c.OnHTML("#builds > div:nth-child(2) > div.container.main.pt-3 > table > tbody > tr:nth-child(18) > td > table > tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr > th", func(i int, ee *colly.HTMLElement) {
			if strings.TrimSpace(strings.ToLower(ee.Text)) == "x86_64" {
				// color.Red("x86_64 found")
				downloadURLTr := ee.DOM.Parent().Next()

				td := downloadURLTr.Find("td").Eq(1)

				td.Find("a").Each(func(i int, s *goquery.Selection) {
					if strings.TrimSpace(strings.ToLower(s.Text())) == "download" {
						href, exists := s.Attr("href")
						if exists {
							x64RpmDownloadURL = href
						}
					}
				})
			}

			if strings.TrimSpace(strings.ToLower(ee.Text)) == "aarch64" {
				downloadURLTr := ee.DOM.Parent().Next()

				td := downloadURLTr.Find("td").Eq(1)

				td.Find("a").Each(func(i int, s *goquery.Selection) {
					if strings.TrimSpace(strings.ToLower(s.Text())) == "download" {
						href, exists := s.Attr("href")
						if exists {
							aarch64RpmDownloadURL = href
						}
					}
				})
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		color.Red("访问出错, url: %s, err: %v\n", r.Request.URL, err)
	})

	return func(url string) (string, string) {
		if err := c.Visit(url); err != nil {
			return "", ""
		}

		return x64RpmDownloadURL, aarch64RpmDownloadURL
	}
}
