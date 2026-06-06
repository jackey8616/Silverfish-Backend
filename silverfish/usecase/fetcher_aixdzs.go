package usecase

import (
	"fmt"
	"regexp"
	"strconv"
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

// aixdzsChapterRe matches /read/<id>/p<n>.html. Used to discover the chapter
// page range from whatever read-links happen to be on the info page (e.g.
// "立即閱讀" → p2, "最新:" → p43); we then enumerate p1..pMax for the catalog.
var aixdzsChapterRe = regexp.MustCompile(`(/read/\d+/)p(\d+)\.html`)

// GetChapterURL export
func (fa *FetcherAixdzs) GetChapterURL(novel *entity.Novel, index int) *string {
	url := "https://" + novel.DNS + novel.Chapters[index].URL
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
	// aixdzs renders title/author/cover via JS into div.n-text, so plain
	// FetchDoc gives us an empty shell. Use Rod to get the post-render DOM.
	doc, docErr := fa.FetchDocViaRod(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fa.GenerateID(url)
	info, infoErr := fa.FetchNovelInfo(id, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fa.FetchChapterInfo(doc, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	novel := &entity.Novel{
		DNS:      *fa.dns,
		URL:      *url,
		Chapters: chapters,
	}
	novel.SetNovelInfo(info)
	return novel, nil
}

// FetchNovelInfo export
func (fa *FetcherAixdzs) FetchNovelInfo(novelID *string, doc *goquery.Document) (*entity.NovelInfo, error) {
	title := strings.TrimSpace(doc.Find("div.n-text > h1").Text())
	author := strings.TrimSpace(doc.Find("a.bauthor").First().Text())
	coverURL, _ := doc.Find("div.n-img > img").Attr("src")
	// New layout no longer carries an inline synopsis; fall back to title so
	// the field is non-empty without inventing content.
	description := title
	if title == "" || author == "" || coverURL == "" {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, coverURL: %s", title, author, coverURL)
	}

	return &entity.NovelInfo{
		IsEnable:      true,
		NovelID:       *novelID,
		Title:         fa.decoder.ConvertString(title),
		Author:        fa.decoder.ConvertString(author),
		Description:   fa.decoder.ConvertString(description),
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fa *FetcherAixdzs) FetchChapterInfo(doc *goquery.Document, title, url string) []entity.NovelChapter {
	chapters := []entity.NovelChapter{}
	chapterPath := ""
	maxPage := 0
	doc.Find("a[href*='/read/']").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		m := aixdzsChapterRe.FindStringSubmatch(href)
		if len(m) != 3 {
			return
		}
		n, err := strconv.Atoi(m[2])
		if err != nil {
			return
		}
		if chapterPath == "" {
			chapterPath = m[1]
		}
		if n > maxPage {
			maxPage = n
		}
	})
	if chapterPath == "" || maxPage == 0 {
		logrus.Printf("Chapter list empty, title: %s, url: %s", title, url)
		return chapters
	}
	for i := 1; i <= maxPage; i++ {
		chapters = append(chapters, entity.NovelChapter{
			Title: fmt.Sprintf("第%d章", i),
			URL:   fmt.Sprintf("%sp%d.html", chapterPath, i),
		})
	}
	return chapters
}

// UpdateNovelInfo export
func (fa *FetcherAixdzs) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := fa.FetchDocViaRod(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fa.FetchNovelInfo(&novel.NovelID, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
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
	doc, docErr := fa.FetchDocViaRod(url)
	if docErr != nil {
		return nil, docErr
	}

	novelContent, _ := doc.Find("div.content").Html()
	output += fa.decoder.ConvertString(novelContent)

	return fa.Filter(&output), nil
}
