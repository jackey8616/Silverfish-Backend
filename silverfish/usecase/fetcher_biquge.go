package usecase

import (
	"fmt"
	"log"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
)

// FetcherBiquge export
type FetcherBiquge struct {
	Fetcher
	charset string
	decoder mahonia.Decoder
}

// NewFetcherBiquge export
func NewFetcherBiquge(dns string) *FetcherBiquge {
	fb := new(FetcherBiquge)
	fb.NewFetcher(true, &dns)

	fb.charset = "utf-8"
	fb.decoder = mahonia.NewDecoder(fb.charset)
	return fb
}

// GetChapterURL export
func (fb *FetcherBiquge) GetChapterURL(novel *entity.Novel, index int) *string {
	url := "https://" + novel.DNS + novel.Chapters[index].URL
	return &url
}

// IsSplit export
func (fb *FetcherBiquge) IsSplit(doc *goquery.Document) bool {
	return false
}

// Filter export
func (fb *FetcherBiquge) Filter(raw *string) *string {
	return raw
}

// FetchNovelInfo export
func (fb *FetcherBiquge) FetchNovelInfo(url *string) (*entity.Novel, error) {
	doc, docErr := fb.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fb.GenerateID(url)
	title := doc.Find("div[id='info'] > h1").Text()
	author := doc.Find("div[id='info'] > p:nth-of-type(1)").Text()
	author = strings.Split(author, "：")[1]
	description := doc.Find("div[id='intro']").Text()
	coverURL, ok := doc.Find("div[id='fmimg'] > img").Attr("src")
	if title == "" || author == "" || description == "" || !ok {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	chapters := []entity.NovelChapter{}
	doc.Find("div[id='list'] > dl > dd > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if chapterTitle != "" && ok {
			chapters = append(chapters, entity.NovelChapter{
				Title: fb.decoder.ConvertString(chapterTitle),
				URL:   chapterURL,
			})
		} else {
			log.Printf("Chapter missing something, title: %s, url: %s", title, *url)
		}
	})

	return &entity.Novel{
		NovelID:       *id,
		DNS:           *fb.dns,
		Title:         fb.decoder.ConvertString(title),
		Author:        fb.decoder.ConvertString(author),
		Description:   fb.decoder.ConvertString(description),
		URL:           *url,
		Chapters:      chapters,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// UpdateNovelInfo export
func (fb *FetcherBiquge) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := fb.FetchDoc(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}

	chapters := []entity.NovelChapter{}
	doc.Find("div[id='list'] > dl > dd > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if chapterTitle != "" && ok {
			chapters = append(chapters, entity.NovelChapter{
				Title: fb.decoder.ConvertString(chapterTitle),
				URL:   chapterURL,
			})
		} else {
			log.Printf("Chapter missing something, title: %s, url: %s", novel.Title, novel.URL)
		}
	})

	novel.Chapters = chapters
	novel.LastCrawlTime = time.Now()
	return novel, nil
}

// FetchNovelChapter export
func (fb *FetcherBiquge) FetchNovelChapter(novel *entity.Novel, index int) (*string, error) {
	url := fb.GetChapterURL(novel, index)
	doc, docErr := fb.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	novelContent, _ := doc.Find("div[id='content']").Html()
	output := fb.decoder.ConvertString(novelContent)

	return fb.Filter(&output), nil
}
