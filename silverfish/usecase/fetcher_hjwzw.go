package usecase

import (
	"fmt"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
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
	index := strings.Index(str, "</p>")
	if index == -1 {
		return &str
	}
	str = str[index+4:]
	str = strings.Replace(str, "讀好書,請記住讀書客唯一地址()</p>", "</p>", 1)
	str = strings.Replace(str, "緊張時放松自己，煩惱時安慰自己，開心時別忘了祝福自己!", "", -1)
	str = strings.Replace(str, "<p>\n</p>", "", -1)
	str = strings.Replace(str, "<p>\n<br/></p>", "", -1)
	return &str
}

// CrawlNovel export
func (fh *FetcherHjwzw) CrawlNovel(url *string) (*entity.Novel, error) {
	doc, docErr := fh.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fh.GenerateId(url)
	info, infoErr := fh.FetchNovelInfo(id, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapterURL := strings.Replace(*url, "Book", "Book/Chapter", 1)
	doc, docErr = fh.FetchDoc(&chapterURL)
	if docErr != nil {
		return nil, docErr
	}
	chapters := fh.FetchChapterInfo(doc, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	novel := &entity.Novel{
		DNS:      *fh.dns,
		URL:      *url,
		Chapters: chapters,
	}
	novel.SetNovelInfo(info)
	return novel, nil
}

// FetchNovelInfo export
func (fh *FetcherHjwzw) FetchNovelInfo(novelId *string, doc *goquery.Document) (*entity.NovelInfo, error) {
	title, ok0 := doc.Find("meta[property='og:novel:book_name']").Attr("content")
	author, ok1 := doc.Find("meta[property='og:novel:author']").Attr("content")
	description, ok2 := doc.Find("meta[property='og:description']").Attr("content")
	coverURL, ok3 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	return &entity.NovelInfo{
		IsEnable:      true,
		NovelId:       *novelId,
		Title:         title,
		Author:        author,
		Description:   description,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fh *FetcherHjwzw) FetchChapterInfo(doc *goquery.Document, title, url string) []entity.NovelChapter {
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
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})

	return chapters
}

// UpdateNovelInfo export
func (fh *FetcherHjwzw) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := fh.FetchDoc(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fh.FetchNovelInfo(&novel.NovelId, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapterURL := strings.Replace(novel.URL, "Book", "Book/Chapter", 1)
	doc, docErr = fh.FetchDoc(&chapterURL)
	if docErr != nil {
		return nil, docErr
	}
	chapters := fh.FetchChapterInfo(doc, info.Title, novel.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	info.LastCrawlTime = time.Now()
	novel.SetNovelInfo(info)
	novel.Chapters = chapters
	return novel, nil
}

// FetchNovelChapter export
func (fh *FetcherHjwzw) FetchNovelChapter(novel *entity.Novel, index int) (*string, error) {
	url := fh.GetChapterURL(novel, index)
	doc, docErr := fh.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	anchor := doc.Find("a[title='" + novel.Title + "']")
	novelDiv := anchor.Parent().Parent()
	novelContent, _ := novelDiv.Html()

	return fh.Filter(&novelContent), nil
}
