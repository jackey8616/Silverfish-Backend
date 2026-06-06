package usecase

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// FetcherUukanshu export
type FetcherUukanshu struct {
	Fetcher
}

// NewFetcherUukanshu export
func NewFetcherUukanshu(dns string) *FetcherUukanshu {
	fu := new(FetcherUukanshu)
	fu.NewFetcher(true, &dns)
	return fu
}

// uukanshuChapterRe matches /book/<bookID>/<chapterID>.html relative URLs
// from chapter list anchors, so we can ignore the breadcrumb / promo links
// that share the same dd ancestor chain on neighboring containers.
var uukanshuChapterRe = regexp.MustCompile(`^/book/\d+/\d+\.html$`)

// uukanshuAdScriptRe scrubs the inline `loadAdv(...)` ad injections that
// sit between paragraphs inside readcotent.
var uukanshuAdScriptRe = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)

// GetChapterURL export
func (fu *FetcherUukanshu) GetChapterURL(novel *entity.Novel, index int) *string {
	url := "https://" + novel.DNS + novel.Chapters[index].URL
	return &url
}

// IsSplit export
func (fu *FetcherUukanshu) IsSplit(doc *goquery.Document) bool {
	return false
}

// Filter export
func (fu *FetcherUukanshu) Filter(raw *string) *string {
	str := uukanshuAdScriptRe.ReplaceAllString(*raw, "")
	return &str
}

// CrawlNovel export
func (fu *FetcherUukanshu) CrawlNovel(url *string) (*entity.Novel, error) {
	// uukanshu.cc sits behind Cloudflare's "Just a moment..." challenge,
	// so plain HTTP returns the interstitial. Drive everything through Rod
	// to get the real post-challenge DOM.
	doc, docErr := fu.FetchDocViaRod(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fu.GenerateID(url)
	info, infoErr := fu.FetchNovelInfo(id, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fu.FetchChapterInfo(doc, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	novel := &entity.Novel{
		DNS:      *fu.dns,
		URL:      *url,
		Chapters: chapters,
	}
	novel.SetNovelInfo(info)
	return novel, nil
}

// FetchNovelInfo export
func (fu *FetcherUukanshu) FetchNovelInfo(novelID *string, doc *goquery.Document) (*entity.NovelInfo, error) {
	title, ok0 := doc.Find("meta[property='og:novel:book_name']").Attr("content")
	author, ok1 := doc.Find("meta[property='og:novel:author']").Attr("content")
	description, ok2 := doc.Find("meta[property='og:description']").Attr("content")
	coverURL, ok3 := doc.Find("meta[property='og:image']").Attr("content")
	if !ok0 || !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	return &entity.NovelInfo{
		IsEnable:      true,
		NovelID:       *novelID,
		Title:         strings.TrimSpace(title),
		Author:        strings.TrimSpace(author),
		Description:   strings.TrimSpace(description),
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fu *FetcherUukanshu) FetchChapterInfo(doc *goquery.Document, title, url string) []entity.NovelChapter {
	chapters := []entity.NovelChapter{}
	// The info page also carries a "latest chapter" promo link with the same
	// href shape, so we filter to anchors inside <dd> (the actual list rows)
	// and validate the URL pattern.
	doc.Find("dd > a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || !uukanshuChapterRe.MatchString(href) {
			return
		}
		chapterTitle := strings.TrimSpace(s.Text())
		if chapterTitle == "" {
			return
		}
		chapters = append(chapters, entity.NovelChapter{
			Title: chapterTitle,
			URL:   href,
		})
	})
	if len(chapters) == 0 {
		logrus.Printf("Chapter list empty, title: %s, url: %s", title, url)
	}
	return chapters
}

// UpdateNovelInfo export
func (fu *FetcherUukanshu) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := fu.FetchDocViaRod(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fu.FetchNovelInfo(&novel.NovelID, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fu.FetchChapterInfo(doc, info.Title, novel.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	info.LastCrawlTime = time.Now()
	novel.SetNovelInfo(info)
	novel.Chapters = chapters
	return novel, nil
}

// FetchNovelChapter export
func (fu *FetcherUukanshu) FetchNovelChapter(novel *entity.Novel, index int) (*string, error) {
	url := fu.GetChapterURL(novel, index)
	doc, docErr := fu.FetchDocViaRod(url)
	if docErr != nil {
		return nil, docErr
	}

	content, _ := doc.Find("div.readcotent").Html()
	return fu.Filter(&content), nil
}
