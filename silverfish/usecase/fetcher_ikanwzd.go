package usecase

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// FetcherIkanwzd export
type FetcherIkanwzd struct {
	Fetcher
}

// NewFetcherIkanwzd export
func NewFetcherIkanwzd(dns string) *FetcherIkanwzd {
	fi := new(FetcherIkanwzd)
	fi.NewFetcher(true, &dns)
	return fi
}

// GetChapterURL export
func (fi *FetcherIkanwzd) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "https://" + comic.DNS + chapterURL
	return &url
}

// CrawlComic export
func (fi *FetcherIkanwzd) CrawlComic(url *string) (*entity.Comic, error) {
	doc, docErr := fi.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fi.GenerateID(url)
	info, infoErr := fi.FetchComicInfo(id, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fi.FetchChapterInfo(doc, nil, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic := &entity.Comic{
		DNS:      *fi.dns,
		URL:      *url,
		Chapters: chapters,
	}
	comic.SetComicInfo(info)
	return comic, nil
}

// FetchComicInfo export
func (fi *FetcherIkanwzd) FetchComicInfo(comicID *string, doc *goquery.Document, cookie []*http.Cookie) (*entity.ComicInfo, error) {
	title := doc.Find("div.info > h1").Text()
	author := doc.Find("div.info > p.subtitle:nth-of-type(2)").Text()
	author = strings.Replace(author, "作者：", "", -1)
	description := doc.Find("div.info > p.content").Text()
	coverURL, ok := doc.Find("div.banner_detail_form > div.cover > img").Attr("src")
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
func (fi *FetcherIkanwzd) FetchChapterInfo(doc *goquery.Document, cookie []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	browser := fi.GenerateRodBrowser()
	defer browser.MustClose()

	page := browser.MustConnect().MustPage(url)
	page.Race().Element("ul.detail-list-select").MustHandle(func(e *rod.Element) {
		eles, _ := e.Elements("li")
		for i := 0; i < len(eles); i++ {
			el, _ := eles[i].Element("a")
			chapterTitle := el.MustText()
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
func (fi *FetcherIkanwzd) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := fi.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fi.FetchComicInfo(&comic.ComicID, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fi.FetchChapterInfo(doc, nil, comic.Title, comic.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic.LastCrawlTime = time.Now()
	comic.SetComicInfo(info)
	comic.Chapters = chapters
	return comic, nil
}

// FetchComicChapter export
func (fi *FetcherIkanwzd) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	url := fi.GetChapterURL(comic, comic.Chapters[index].URL)
	browser := fi.GenerateRodBrowser()
	defer browser.MustClose()

	page := browser.MustConnect().MustPage(*url)
	els, _ := page.Elements("img.lazy")
	for i := 0; i < len(els); i++ {
		comicURLs = append(comicURLs, *els[i].MustAttribute("data-original"))
	}
	return comicURLs, nil
}
