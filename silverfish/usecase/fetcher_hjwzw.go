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
	str = str[strings.Index(str, "</p>")+4:]
	str = strings.Replace(str, "讀好書,請記住讀書客唯一地址()</p>", "</p>", 1)
	str = strings.Replace(str, "緊張時放松自己，煩惱時安慰自己，開心時別忘了祝福自己!", "", -1)
	str = strings.Replace(str, "<p>\n</p>", "", -1)
	str = strings.Replace(str, "<p>\n<br/></p>", "", -1)
	return &str
}

// FetchNovelInfo export
func (fh *FetcherHjwzw) FetchNovelInfo(url *string) *entity.Novel {
	doc := fh.FetchDoc(url)

	id := fh.GenerateID(url)
	title, ok0 := doc.Find("meta[property='og:novel:book_name']").Attr("content")
	author, ok1 := doc.Find("meta[property='og:novel:author']").Attr("content")
	description, ok2 := doc.Find("meta[property='og:description']").Attr("content")
	coverURL, ok3 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 || !ok3 {
		log.Printf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
		return nil
	}
	chapterURL := strings.Replace(*url, "Book", "Book/Chapter", 1)
	doc = fh.FetchDoc(&chapterURL)

	chapters := []entity.NovelChapter{}
	doc.Find("div#tbchapterlist > table > tbody > tr > td > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append(chapters, entity.NovelChapter{
				Title: chapterTitle,
				URL:   chapterURL,
			})
		} else {
			log.Printf("Chapter missing something, title: %s, url: %s", title, *url)
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

// UpdateNovelInfo export
func (fh *FetcherHjwzw) UpdateNovelInfo(novel *entity.Novel) *entity.Novel {
	chapterURL := strings.Replace(novel.URL, "Book", "Book/Chapter", 1)
	doc := fh.FetchDoc(&chapterURL)

	chapters := []entity.NovelChapter{}
	doc.Find("div#tbchapterlist > table > tbody > tr > td > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append(chapters, entity.NovelChapter{
				Title: chapterTitle,
				URL:   chapterURL,
			})
		} else {
			log.Printf("Chapter missing something, title: %s, url: %s", novel.Title, novel.URL)
		}
	})

	novel.Chapters = chapters
	novel.LastCrawlTime = time.Now()
	return novel
}

// FetchNovelChapter export
func (fh *FetcherHjwzw) FetchNovelChapter(novel *entity.Novel, index int) *string {
	url := fh.GetChapterURL(novel, index)
	doc := fh.FetchDoc(url)

	anchor := doc.Find("a[title='" + novel.Title + "']")
	novelDiv := anchor.Parent().Parent()
	novelContent, _ := novelDiv.Html()

	return fh.Filter(&novelContent)
}
