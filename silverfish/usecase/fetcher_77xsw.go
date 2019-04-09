package usecase

import (
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

// GetChapterURL export
func (f7 *Fetcher77xsw) GetChapterURL(novel *entity.Novel, index int) *string {
	url := novel.URL + novel.Chapters[index].URL
	return &url
}

// IsSplit export
func (f7 *Fetcher77xsw) IsSplit(doc *goquery.Document) bool {
	el := doc.Find("h1.readTitle > small").Text()
	return el == "(1/2)"
}

// Filter export
func (f7 *Fetcher77xsw) Filter(raw *string) *string {
	str := *raw
	str = strings.Replace(str, "一秒记住【千千小说网 www.77xsw.la】，更新快，无弹窗，免费读！", "", -1)
	str = strings.Replace(str, "本章未完，点击下一页继续阅读", "", -1)
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
	description, ok2 := doc.Find("meta[property='og:description']").Attr("content")
	coverURL, ok3 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 || !ok3 {
		log.Printf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
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
			log.Printf("Chapter missing something, title: %s, url: %s", title, *url)
		}
	})

	return &entity.Novel{
		NovelID:       *id,
		DNS:           *f7.dns,
		Title:         f7.decoder.ConvertString(title),
		Author:        f7.decoder.ConvertString(author),
		Description:   f7.decoder.ConvertString(description),
		URL:           *url,
		Chapters:      chapters,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}
}

// FetchChapter export
func (f7 *Fetcher77xsw) FetchChapter(novel *entity.Novel, index int) *string {
	url := f7.GetChapterURL(novel, index)
	output := ""
	doc := f7.FetchDoc(url)

	novelContent := doc.Find("div#htmlContent").Text()
	output += f7.decoder.ConvertString(novelContent)

	if f7.IsSplit(doc) {
		url2 := strings.Replace(*url, ".html", "_2.html", 1)
		doc = f7.FetchDoc(&url2)
		novelContent = doc.Find("div#htmlContent").Text()
		output += f7.decoder.ConvertString(novelContent)
	}

	return f7.Filter(&output)
}
