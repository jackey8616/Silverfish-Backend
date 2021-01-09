package usecase

import (
	"fmt"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/sirupsen/logrus"
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

// CrawlNovel export
func (f7 *Fetcher77xsw) CrawlNovel(url *string) (*entity.Novel, error) {
	doc, docErr := f7.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := f7.GenerateID(url)
	info, infoErr := f7.FetchNovelInfo(id, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := f7.FetchChapterInfo(doc, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	novel := &entity.Novel{
		DNS:      *f7.dns,
		URL:      *url,
		Chapters: chapters,
	}
	novel.SetNovelInfo(info)
	return novel, nil
}

// FetchNovelInfo export
func (f7 *Fetcher77xsw) FetchNovelInfo(novelID *string, doc *goquery.Document) (*entity.NovelInfo, error) {
	title, ok0 := doc.Find("meta[property='og:title']").Attr("content")
	author, ok1 := doc.Find("meta[property='og:novel:author']").Attr("content")
	description, ok2 := doc.Find("meta[property='og:description']").Attr("content")
	coverURL, ok3 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	return &entity.NovelInfo{
		IsEnable:      true,
		NovelID:       *novelID,
		Title:         f7.decoder.ConvertString(title),
		Author:        f7.decoder.ConvertString(author),
		Description:   f7.decoder.ConvertString(description),
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (f7 *Fetcher77xsw) FetchChapterInfo(doc *goquery.Document, title, url string) []entity.NovelChapter {
	chapters := []entity.NovelChapter{}

	doc.Find("div#list-chapterAll > dl > dd > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle, ok0 := s.Attr("title")
		chapterURL, ok1 := s.Attr("href")
		if ok0 && ok1 {
			chapters = append(chapters, entity.NovelChapter{
				Title: f7.decoder.ConvertString(chapterTitle),
				URL:   chapterURL,
			})
		} else {
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})

	return chapters
}

// UpdateNovelInfo export
func (f7 *Fetcher77xsw) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := f7.FetchDoc(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}
	info, infoErr := f7.FetchNovelInfo(&novel.NovelID, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := f7.FetchChapterInfo(doc, novel.Title, novel.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	info.LastCrawlTime = time.Now()
	novel.SetNovelInfo(info)
	novel.Chapters = chapters
	return novel, nil
}

// FetchNovelChapter export
func (f7 *Fetcher77xsw) FetchNovelChapter(novel *entity.Novel, index int) (*string, error) {
	url := f7.GetChapterURL(novel, index)
	output := ""
	doc, docErr := f7.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	novelContent := doc.Find("div#htmlContent").Text()
	output += f7.decoder.ConvertString(novelContent)

	if f7.IsSplit(doc) {
		url2 := strings.Replace(*url, ".html", "_2.html", 1)
		doc, docErr = f7.FetchDoc(&url2)
		if docErr != nil {
			return nil, docErr
		}
		novelContent = doc.Find("div#htmlContent").Text()
		output += f7.decoder.ConvertString(novelContent)
	}

	return f7.Filter(&output), nil
}
