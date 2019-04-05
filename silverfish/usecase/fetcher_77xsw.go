package usecase

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/jackey8616/Silverfish-backend/silverfish/entity"
)

// Fetcher77xsw export
type Fetcher77xsw struct {
	Fetcher
	charset string
	decoder mahonia.Decoder
}

// NewFetcher77xsw export
func NewFetcher77xsw(dns string) *Fetcher77xsw {
	f7 := new(Fetcher77xsw)
	f7.NewFetcher(false, &dns)

	f7.charset = "gbk"
	f7.decoder = mahonia.NewDecoder(f7.charset)
	return f7
}

// IsSplit export
func (f7 *Fetcher77xsw) IsSplit(doc *goquery.Document) bool {
	el := doc.Find("h1.readTitle > small").Text()
	return el == "(1/2)"
}

// Filter export
func (f7 *Fetcher77xsw) Filter(raw *string) *string {
	str := *raw
	str = strings.Replace(str, "\n                一秒记住【千千小说网 www.77xsw.la】，更新快，无弹窗，免费读！\n                ", "", 1)
	str = strings.Replace(str, "\n                -->>本章未完，点击下一页继续阅读\n            \n                一秒记住【千千小说网 www.77xsw.la】，更新快，无弹窗，免费读！\n                ", "", 1)
	str = strings.Replace(str, "聽聽聽聽", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	str = strings.Replace(str, "&nbsp;聽聽聽", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	str = strings.Replace(str, "聽&nbsp;聽聽", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	str = strings.Replace(str, "聽聽&nbsp;聽", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	str = strings.Replace(str, "聽聽聽&nbsp;", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	str = strings.Replace(str, "\n", "<br/>", -1)
	return &str
}

// FetchNovelInfo export
func (f7 *Fetcher77xsw) FetchNovelInfo(url *string) *entity.Novel {
	doc := f7.FetchDoc(url)

	id := f7.GenerateID(url)
	title, ok0 := doc.Find("meta[property='og:title']").Attr("content")
	author, ok1 := doc.Find("meta[property='og:novel:author']").Attr("content")
	decription, ok2 := doc.Find("meta[property='og:description']").Attr("content")
	coverURL, ok3 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 || !ok3 {
		log.Fatal(fmt.Sprintf("Something missing, title: %t, author: %t, description: %t, coverURL: %t", ok0, ok1, ok2, ok3))
		return nil
	}

	chapters := []entity.Chapter{}
	doc.Find("div#list-chapterAll > dl > dd > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle, ok0 := s.Attr("title")
		chapterURL, ok1 := s.Attr("href")
		if ok0 && ok1 {
			chapters = append(chapters, entity.Chapter{
				Title: f7.decoder.ConvertString(chapterTitle),
				URL:   chapterURL,
			})
		} else {
			log.Fatal(fmt.Sprintf("Chapter missing something, title: %t, url: %t", ok0, ok1))
		}
	})

	return &entity.Novel{
		NovelID:       *id,
		DNS:           *f7.dns,
		Title:         f7.decoder.ConvertString(title),
		Author:        f7.decoder.ConvertString(author),
		Description:   f7.decoder.ConvertString(decription),
		URL:           *url,
		Chapters:      chapters,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}
}

// FetchChapter export
func (f7 *Fetcher77xsw) FetchChapter(novel *entity.Novel, index int) *string {
	url := novel.URL + novel.Chapters[index].URL
	output := ""
	doc := f7.FetchDoc(&url)

	novelContent := doc.Find("div#htmlContent").Text()
	output += f7.decoder.ConvertString(novelContent)

	if f7.IsSplit(doc) {
		url = strings.Replace(url, ".html", "_2.html", 1)
		doc = f7.FetchDoc(&url)
		novelContent = doc.Find("div#htmlContent").Text()
		output += f7.decoder.ConvertString(novelContent)
	}

	return f7.Filter(&output)
}
