package usecase

import (
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jackey8616/Silverfish-backend/silverfish/entity"
)

// FetcherHjwzw export
type FetcherHjwzw struct {
	Fetcher
}

// NewFetcherHjwzw export
func NewFetcherHjwzw(dns string) *FetcherHjwzw {
	fh := new(FetcherHjwzw)
	fh.NewFetcher(true, &dns)
	return fh
}

// GetChapterURL export
func (fh *FetcherHjwzw) GetChapterURL(novel *entity.Novel, index int) *string {
	url := "https://" + novel.DNS + novel.Chapters[index].URL
	return &url
}

// IsSplit export
func (fh *FetcherHjwzw) IsSplit(doc *goquery.Document) bool {
	return false
}

// Filter export
func (fh *FetcherHjwzw) Filter(raw *string) *string {
	str := *raw
	str = strings.Replace(str, "請記住本站域名: 黃金屋", "", -1)
	str = strings.Replace(str, "哥們，別忘了收藏！", "", -1)
	str = strings.Replace(str, "\n", "", 3)
	str = strings.Replace(str, "\n", "<br/>", -1)
	return &str
}

// FetchNovelInfo export
func (fh *FetcherHjwzw) FetchNovelInfo(url *string) *entity.Novel {
	doc := fh.FetchDoc(url)

	id := fh.GenerateID(url)
	title := doc.Find("#form1 > div:nth-child(4) > table:nth-child(10) > tbody:nth-child(1) > tr:nth-child(1) > td:nth-child(1) > h1:nth-child(1)").Text()
	author := doc.Find("#form1 > div:nth-child(4) > table:nth-child(10) > tbody:nth-child(1) > tr:nth-child(2) > td:nth-child(1) > a:nth-child(1)").Text()
	description, ok2 := doc.Find("meta[name='Description']").Attr("content")
	coverURL := strings.Replace(*url, "Book/Chapter", "images/id", 1) + ".jpg"
	if title == "" || author == "" || !ok2 {
		log.Printf("Something missing, title: %s, author: %s, description: %s", title, author, description)
		return nil
	}

	chapters := []entity.Chapter{}
	doc.Find("div#tbchapterlist > table > tbody > tr > td > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append(chapters, entity.Chapter{
				Title: chapterTitle,
				URL:   chapterURL,
			})
		} else {
			log.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})

	return &entity.Novel{
		NovelID:       *id,
		DNS:           *fh.dns,
		Title:         title,
		Author:        author,
		Description:   description,
		URL:           *url,
		Chapters:      chapters,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}
}

// FetchChapter export
func (fh *FetcherHjwzw) FetchChapter(novel *entity.Novel, index int) *string {
	url := fh.GetChapterURL(novel, index)
	doc := fh.FetchDoc(url)

	anchor := doc.Find("#form1 > table > tbody > tr > td > div > p > a")
	novelDiv := anchor.Parent().Parent()
	novelContent := novelDiv.Text()

	return fh.Filter(&novelContent)
}
