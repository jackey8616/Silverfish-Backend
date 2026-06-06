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

// FetcherTtkan export
type FetcherTtkan struct {
	Fetcher
}

// NewFetcherTtkan export
func NewFetcherTtkan(dns string) *FetcherTtkan {
	ft := new(FetcherTtkan)
	ft.NewFetcher(true, &dns)
	return ft
}

// ttkanChapterRe matches /novel/pagea/<slug>_<n>.html links. Slug may
// contain `_` itself (e.g. books with a series prefix like
// `quanmindahanghai_wokaiju...`); regex is greedy so `_\d+\.html$`
// anchors to the trailing chapter index. Also rejects the unrendered
// amp-mustache template anchor (href ends in `{{chapter_id}}`).
var ttkanChapterRe = regexp.MustCompile(`^/novel/pagea/[a-z0-9_-]+_\d+\.html$`)

// GetChapterURL export
func (ft *FetcherTtkan) GetChapterURL(novel *entity.Novel, index int) *string {
	url := "https://" + novel.DNS + novel.Chapters[index].URL
	return &url
}

// IsSplit export
func (ft *FetcherTtkan) IsSplit(doc *goquery.Document) bool {
	return false
}

// Filter export — chapter HTML scrubbing happens via goquery in
// FetchNovelChapter, so this is a pass-through.
func (ft *FetcherTtkan) Filter(raw *string) *string {
	return raw
}

// CrawlNovel export
func (ft *FetcherTtkan) CrawlNovel(url *string) (*entity.Novel, error) {
	doc, docErr := ft.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := ft.GenerateID(url)
	info, infoErr := ft.FetchNovelInfo(id, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := ft.FetchChapterInfo(doc, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	novel := &entity.Novel{
		DNS:      *ft.dns,
		URL:      *url,
		Chapters: chapters,
	}
	novel.SetNovelInfo(info)
	return novel, nil
}

// FetchNovelInfo export — ttkan's Nuxt SSR emits `name="og:..."`
// rather than `property="og:..."`, which is why these selectors differ
// from fetcher_hjwzw.go.
func (ft *FetcherTtkan) FetchNovelInfo(novelID *string, doc *goquery.Document) (*entity.NovelInfo, error) {
	title, ok0 := doc.Find("meta[name='og:novel:book_name']").Attr("content")
	author, ok1 := doc.Find("meta[name='og:novel:author']").Attr("content")
	description, ok2 := doc.Find("meta[name='og:description']").Attr("content")
	coverURL, ok3 := doc.Find("meta[name='og:image']").Attr("content")
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
func (ft *FetcherTtkan) FetchChapterInfo(doc *goquery.Document, title, url string) []entity.NovelChapter {
	chapters := []entity.NovelChapter{}
	doc.Find("div.full_chapters a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || !ttkanChapterRe.MatchString(href) {
			return
		}
		chapterTitle := strings.TrimSpace(s.Text())
		if chapterTitle == "" {
			// `<slug>_1.html` consistently renders with empty link text;
			// it's a preface placeholder, not a real chapter.
			return
		}
		chapters = append(chapters, entity.NovelChapter{
			Title: chapterTitle,
			URL:   href,
		})
	})

	return chapters
}

// UpdateNovelInfo export
func (ft *FetcherTtkan) UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error) {
	doc, docErr := ft.FetchDoc(&novel.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := ft.FetchNovelInfo(&novel.NovelID, doc)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := ft.FetchChapterInfo(doc, info.Title, novel.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	info.LastCrawlTime = time.Now()
	novel.SetNovelInfo(info)
	novel.Chapters = chapters
	return novel, nil
}

// FetchNovelChapter export
func (ft *FetcherTtkan) FetchNovelChapter(novel *entity.Novel, index int) (*string, error) {
	url := ft.GetChapterURL(novel, index)
	doc, docErr := ft.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	content := doc.Find("div.content").First()
	// Strip non-text injections: bookmark anchor at the top, the
	// mid-content `<center><div class="mobadsq"></div></center>` ad
	// block, and amp-* / script nodes.
	content.Find("a.anchor_bookmark, center, amp-img, amp-analytics, script").Remove()
	html, _ := content.Html()
	return ft.Filter(&html), nil
}
