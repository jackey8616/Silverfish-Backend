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

// FetcherAixdzs export
type FetcherAixdzs struct {
	Fetcher
	charset string
	decoder mahonia.Decoder
}

// NewFetcherAixdzs export
func NewFetcherAixdzs(dns string) *FetcherAixdzs {
	fa := new(FetcherAixdzs)
	fa.NewFetcher(false, &dns)

	fa.charset = "utf-8"
	fa.decoder = mahonia.NewDecoder(fa.charset)
	return fa
}

// GetChapterURL export
func (fa *FetcherAixdzs) GetChapterURL(novel *entity.Novel, index int) *string {
	url := novel.URL + novel.Chapters[index].URL
	url = strings.Replace(url, "/d/", "/read/", 1)
	return &url
}

// IsSplit export
func (fa *FetcherAixdzs) IsSplit(doc *goquery.Document) bool {
	return false
}

// Filter export
func (fa *FetcherAixdzs) Filter(raw *string) *string {
	return raw
}

// FetchNovelInfo export
func (fa *FetcherAixdzs) FetchNovelInfo(url *string) (*entity.Novel, error) {
	doc, docErr := fa.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fa.GenerateID(url)
	title, ok0 := doc.Find("meta[property='og:title']").Attr("content")
	author, ok1 := doc.Find("meta[property='og:novel:author']").Attr("content")
	description, ok2 := doc.Find("meta[property='og:description']").Attr("content")
	coverURL, ok3 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	chaptersURL := strings.Replace(*url, "/d/", "/read/", 1)
	doc, docErr = fa.FetchDoc(&chaptersURL)
	if docErr != nil {
		return nil, docErr
	}
	chapters := []entity.NovelChapter{}
	doc.Find("div.catalog > ul > li.chapter > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append(chapters, entity.NovelChapter{
				Title: fa.decoder.ConvertString(chapterTitle),
				URL:   chapterURL,
			})
		} else {
			log.Printf("Chapter missing something, title: %s, url: %s", title, *url)
		}
	})

	return &entity.Novel{
		NovelID:       *id,
		DNS:           *fa.dns,
		Title:         fa.decoder.ConvertString(title),
		Author:        fa.decoder.ConvertString(author),
		Description:   fa.decoder.ConvertString(description),
		URL:           *url,
		Chapters:      chapters,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// UpdateNovelInfo export
func (fa *FetcherAixdzs) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := fa.FetchDoc(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}

	chapters := []entity.NovelChapter{}
	doc.Find("div.catalog > ul > li.chapter > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append(chapters, entity.NovelChapter{
				Title: fa.decoder.ConvertString(chapterTitle),
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
func (fa *FetcherAixdzs) FetchNovelChapter(novel *entity.Novel, index int) (*string, error) {
	url := fa.GetChapterURL(novel, index)
	output := ""
	doc, docErr := fa.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	novelContent, _ := doc.Find("div.content").Html()
	fmt.Println(novelContent)
	output += fa.decoder.ConvertString(novelContent)

	return fa.Filter(&output), nil
}
