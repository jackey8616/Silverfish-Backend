package usecase

import (
	"fmt"
	"net/http"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// FetcherJmd8 export
type FetcherJmd8 struct {
	Fetcher
}

// NewFetcherJmd8 export
func NewFetcherJmd8(dns string) *FetcherJmd8 {
	fj := new(FetcherJmd8)
	fj.NewFetcher(true, &dns)
	return fj
}

// GetChapterURL export
func (fj *FetcherJmd8) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "https://" + comic.DNS + chapterURL
	return &url
}

// CrawlComic export
func (fj *FetcherJmd8) CrawlComic(url *string) (*entity.Comic, error) {
	doc, docErr := fj.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fj.GenerateID(url)
	info, infoErr := fj.FetchComicInfo(id, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fj.FetchChapterInfo(doc, nil, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic := &entity.Comic{
		DNS:      *fj.dns,
		URL:      *url,
		Chapters: chapters,
	}
	comic.SetComicInfo(info)
	return comic, nil
}

// FetchComicInfo export
func (fj *FetcherJmd8) FetchComicInfo(comicID *string, doc *goquery.Document, cookie []*http.Cookie) (*entity.ComicInfo, error) {
	title := doc.Find("h1.page-title").Text()
	author := ""
	description := ""
	doc.Find("div.video-info-items").Each(func(i int, p *goquery.Selection) {
		switch i {
		case 1:
			author = p.Find("div.video-info-item > a").Text()
		case 3:
			description = p.Find("div.video-info-content > span").Text()
		}
	})
	coverURL, ok := doc.Find("div.module-item-pic > img").Attr("src")
	if title == "" || author == "" || description == "" || !ok {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	return &entity.ComicInfo{
		IsEnable:      false,
		ComicID:       *comicID,
		Title:         title,
		Author:        author,
		Description:   description,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fj *FetcherJmd8) FetchChapterInfo(doc *goquery.Document, cookie []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	browser := rod.New()
	defer browser.MustClose()

	page := browser.MustConnect().MustPage(url)
	page.Race().Element("div.module-blocklist > div").MustHandle(func(e *rod.Element) {
		eles, _ := e.Elements("a.detail-write")
		for i := 0; i < len(eles); i++ {
			el := eles[i]
			titleEl, _ := el.Element("span")
			chapterTitle := titleEl.MustText()
			chapterURL := el.MustAttribute("href")
			chapters = append(chapters, entity.ComicChapter{
				Title:    chapterTitle,
				URL:      *chapterURL,
				ImageURL: []string{},
			})
		}
	}).MustDo()
	return chapters
}

// UpdateComicInfo export
func (fj *FetcherJmd8) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := fj.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fj.FetchComicInfo(&comic.ComicID, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fj.FetchChapterInfo(doc, nil, comic.Title, comic.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic.LastCrawlTime = time.Now()
	comic.SetComicInfo(info)
	comic.Chapters = chapters
	return comic, nil
}

// FetchComicChapter export
func (fj *FetcherJmd8) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	browser := rod.New()
	defer browser.MustClose()

	url := fj.GetChapterURL(comic, comic.Chapters[index].URL)
	page := browser.MustConnect().MustPage(*url)
	page.Race().Element("main#main > div > center > div").MustHandle(func(e *rod.Element) {
		eles, _ := e.Elements("img")
		for i := 0; i < len(eles); i++ {
			comicURLs = append(comicURLs, *eles[i].MustAttribute("data-original"))
		}
	}).MustDo()
	return comicURLs, nil
}
