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

// CrawlNovel export
func (fa *FetcherAixdzs) CrawlNovel(url *string) (*entity.Novel, error) {
	doc, docErr := fa.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fa.GenerateID(url)
	description, ok := doc.Find("meta[property='og:description']").Attr("content")
	if !ok {
		return nil, fmt.Errorf("Something missing, description: %s", description)
	}

	info, infoErr := fa.FetchNovelInfo(id, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chaptersURL := strings.Replace(*url, "/d/", "/read/", 1)
	doc, docErr = fa.FetchDoc(&chaptersURL)
	if docErr != nil {
		return nil, docErr
	}
	chapters := fa.FetchChapterInfo(doc, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	novel := &entity.Novel{
		DNS:         *fa.dns,
		Description: fa.decoder.ConvertString(description),
		URL:         *url,
		Chapters:    chapters,
	}
	novel.SetNovelInfo(info)
	return novel, nil
}

// FetchNovelInfo export
func (fa *FetcherAixdzs) FetchNovelInfo(novelID *string, doc *goquery.Document) (*entity.NovelInfo, error) {
	title, ok0 := doc.Find("meta[property='og:title']").Attr("content")
	author, ok1 := doc.Find("meta[property='og:novel:author']").Attr("content")
	coverURL, ok2 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, coverURL: %s", title, author, coverURL)
	}

	return &entity.NovelInfo{
		NovelID:       *novelID,
		Title:         fa.decoder.ConvertString(title),
		Author:        fa.decoder.ConvertString(author),
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fa *FetcherAixdzs) FetchChapterInfo(doc *goquery.Document, title, url string) []entity.NovelChapter {
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
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})

	return chapters
}

// UpdateNovelInfo export
func (fa *FetcherAixdzs) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := fa.FetchDoc(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fa.FetchNovelInfo(&novel.NovelID, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chaptersURL := strings.Replace(novel.URL, "/d/", "/read/", 1)
	doc, docErr = fa.FetchDoc(&chaptersURL)
	if docErr != nil {
		return nil, docErr
	}
	chapters := fa.FetchChapterInfo(doc, info.Title, novel.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	info.LastCrawlTime = time.Now()
	novel.SetNovelInfo(info)
	novel.Chapters = chapters
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
	logrus.Print(novelContent)
	output += fa.decoder.ConvertString(novelContent)

	return fa.Filter(&output), nil
}
