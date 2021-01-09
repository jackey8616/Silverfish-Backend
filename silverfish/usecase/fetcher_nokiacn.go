package usecase

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// FetcherNokiacn export
type FetcherNokiacn struct {
	Fetcher
}

// NewFetcherNokiacn export
func NewFetcherNokiacn(dns string) *FetcherNokiacn {
	fn := new(FetcherNokiacn)
	fn.NewFetcher(true, &dns)
	return fn
}

// GetChapterURL export
func (fn *FetcherNokiacn) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "http://" + comic.DNS + chapterURL
	return &url
}

// CrawlComic export
func (fn *FetcherNokiacn) CrawlComic(url *string) (*entity.Comic, error) {
	doc, docErr := fn.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fn.GenerateID(url)
	info, infoErr := fn.FetchComicInfo(id, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fn.FetchChapterInfo(doc, nil, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic := &entity.Comic{
		DNS:      *fn.dns,
		URL:      *url,
		Chapters: chapters,
	}
	comic.SetComicInfo(info)
	return comic, nil
}

// FetchComicInfo export
func (fn *FetcherNokiacn) FetchComicInfo(comicID *string, doc *goquery.Document, cookies []*http.Cookie) (*entity.ComicInfo, error) {
	title := doc.Find("div.cy_title > h1").Text()
	author := doc.Find("div.cy_xinxi:nth-child(4) > span > a").Text()
	description := doc.Find("p#comic-description").Text()
	coverURL, ok := doc.Find("div.cy_info_cover > a > img").Attr("src")
	// The coverURL is still using 55888, manual replace it.
	coverURL = strings.Replace(coverURL, ":55888", "", 1)
	if title == "" || author == "" || description == "'" || !ok {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	return &entity.ComicInfo{
		IsEnable:      true,
		ComicID:       *comicID,
		Title:         title,
		Author:        author,
		Description:   description,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fn *FetcherNokiacn) FetchChapterInfo(doc *goquery.Document, cookies []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	doc.Find("ul#mh-chapter-list-ol-0 > li > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Find("p").Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append([]entity.ComicChapter{
				entity.ComicChapter{
					Title:    chapterTitle,
					URL:      chapterURL,
					ImageURL: []string{},
				},
			}, chapters...)
		} else {
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})

	return chapters
}

// UpdateComicInfo export
func (fn *FetcherNokiacn) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := fn.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fn.FetchComicInfo(&comic.ComicID, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fn.FetchChapterInfo(doc, nil, comic.Title, comic.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic.LastCrawlTime = time.Now()
	comic.SetComicInfo(info)
	comic.Chapters = chapters
	return comic, nil
}

// FetchComicChapter export
func (fn *FetcherNokiacn) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	firstPageURL := fn.GetChapterURL(comic, comic.Chapters[index].URL)
	doc, docErr := fn.FetchDoc(firstPageURL)
	if docErr != nil {
		return nil, docErr
	}
	firstPage, _ := doc.Html()

	rImages, _ := regexp.Compile(`var qTcms_S_m_murl_e=\".*?\";`)
	base64Code := rImages.FindString(firstPage)
	decode, _ := base64.StdEncoding.DecodeString(base64Code[22 : len(base64Code)-2])
	images := strings.Split(string(decode), "$qingtiandy$")

	for i := 0; i < len(images); i++ {
		imageURL := "http://n.aiwenwo.net" + images[i]
		comicURLs = append(comicURLs, imageURL)
	}
	return comicURLs, nil
}
