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

// FetcherBookbl export
type FetcherBookbl struct {
	Fetcher
	charset string
	decoder mahonia.Decoder
}

// NewFetcherBookbl export
func NewFetcherBookbl(dns string) *FetcherBookbl {
	fb := new(FetcherBookbl)
	fb.NewFetcher(false, &dns)

	fb.charset = "utf-8"
	fb.decoder = mahonia.NewDecoder(fb.charset)
	return fb
}

// GetChapterURL export
func (fb *FetcherBookbl) GetChapterURL(novel *entity.Novel, index int) *string {
	url := fmt.Sprintf(`https://%s%s`, novel.DNS, novel.Chapters[index].URL)
	return &url
}

// IsSplit export
func (fb *FetcherBookbl) IsSplit(doc *goquery.Document) bool {
	return false
}

// Filter export
func (fb *FetcherBookbl) Filter(raw *string) *string {
	return raw
}

// CrawlNovel export
func (fb *FetcherBookbl) CrawlNovel(url *string) (*entity.Novel, error) {
	doc, docErr := fb.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fb.GenerateID(url)
	info, infoErr := fb.FetchNovelInfo(id, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fb.FetchChapterInfo(doc, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	novel := &entity.Novel{
		DNS:      *fb.dns,
		URL:      *url,
		Chapters: chapters,
	}
	novel.SetNovelInfo(info)
	return novel, nil
}

// FetchNovelInfo export
func (fb *FetcherBookbl) FetchNovelInfo(novelID *string, doc *goquery.Document) (*entity.NovelInfo, error) {
	title, ok0 := doc.Find("meta[property='og:title']").Attr("content")
	author, ok1 := doc.Find("meta[property='og:novel:author']").Attr("content")
	description := doc.Find("div.intro").Text()
	description = strings.Replace(description, "簡介:", "", -1)
	coverURL, ok2 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 || description == "" {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	return &entity.NovelInfo{
		IsEnable:      true,
		NovelID:       *novelID,
		Title:         fb.decoder.ConvertString(title),
		Author:        fb.decoder.ConvertString(author),
		Description:   fb.decoder.ConvertString(description),
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fb *FetcherBookbl) FetchChapterInfo(doc *goquery.Document, title, url string) []entity.NovelChapter {
	chapters := []entity.NovelChapter{}

	doc.Find("div.panel-booklist > ul.list-group > li > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle, ok0 := s.Attr("title")
		chapterURL, ok1 := s.Attr("href")
		if ok0 && ok1 {
			chapters = append(chapters, entity.NovelChapter{
				Title: fb.decoder.ConvertString(chapterTitle),
				URL:   chapterURL,
			})
		} else {
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})

	return chapters
}

// UpdateNovelInfo export
func (fb *FetcherBookbl) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := fb.FetchDoc(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}
	info, infoErr := fb.FetchNovelInfo(&novel.NovelID, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fb.FetchChapterInfo(doc, novel.Title, novel.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	info.LastCrawlTime = time.Now()
	novel.SetNovelInfo(info)
	novel.Chapters = chapters
	return novel, nil
}

// FetchNovelChapter export
func (fb *FetcherBookbl) FetchNovelChapter(novel *entity.Novel, index int) (*string, error) {
	url := fb.GetChapterURL(novel, index)
	output := ""
	doc, docErr := fb.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	novelContent, _ := doc.Find("div.content").Html()
	output += fb.decoder.ConvertString(novelContent)

	return fb.Filter(&output), nil
}
