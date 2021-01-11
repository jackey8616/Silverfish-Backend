package usecase

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// FetcherMfhmh export
type FetcherMfhmh struct {
	Fetcher
}

// NewFetcherMfhmh export
func NewFetcherMfhmh(dns string) *FetcherMfhmh {
	fm := new(FetcherMfhmh)
	fm.NewFetcher(true, &dns)
	return fm
}

// GetChapterURL export
func (fm *FetcherMfhmh) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "https://" + comic.DNS + chapterURL
	return &url
}

// CrawlComic export
func (fm *FetcherMfhmh) CrawlComic(url *string) (*entity.Comic, error) {
	doc, docErr := fm.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fm.GenerateID(url)
	info, infoErr := fm.FetchComicInfo(id, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fm.FetchChapterInfo(doc, nil, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic := &entity.Comic{
		DNS:      *fm.dns,
		URL:      *url,
		Chapters: chapters,
	}
	comic.SetComicInfo(info)
	return comic, nil
}

// FetchComicInfo export
func (fm *FetcherMfhmh) FetchComicInfo(comicID *string, doc *goquery.Document, cookie []*http.Cookie) (*entity.ComicInfo, error) {
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
func (fm *FetcherMfhmh) FetchChapterInfo(doc *goquery.Document, cookie []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	browser := rod.New()
	defer browser.MustClose()

	page := browser.MustConnect().MustPage(url)
	page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "iPhone",
	})
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
func (fm *FetcherMfhmh) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := fm.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fm.FetchComicInfo(&comic.ComicID, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fm.FetchChapterInfo(doc, nil, comic.Title, comic.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic.LastCrawlTime = time.Now()
	comic.SetComicInfo(info)
	comic.Chapters = chapters
	return comic, nil
}

// FetchComicChapter export
func (fm *FetcherMfhmh) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	browser := rod.New()
	defer browser.MustClose()

	page := browser.MustConnect().MustPage(comic.Chapters[index].URL)
	page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: "iPhone",
	})
	widthExp, _ := regexp.Compile(`width: .*?;`)
	page.Race().Element("div#cp_img").MustHandle(func(e *rod.Element) {
		eles, _ := e.Elements("img")
		for i := 0; i < len(eles); i++ {
			width := widthExp.FindString(*eles[i].MustAttribute("style"))
			comicURLs = append(comicURLs, *eles[i].MustAttribute("data-original")+"?"+width[7:len(width)-2]+"%")
		}
	}).MustDo()
	return comicURLs, nil
}
