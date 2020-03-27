package usecase

import (
	"encoding/base64"
	"fmt"
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

// FetchComicInfo export
func (fn *FetcherNokiacn) FetchComicInfo(url *string) (*entity.Comic, error) {
	doc, docErr := fn.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fn.GenerateID(url)
	title := doc.Find("div.cy_title > h1").Text()
	author := doc.Find("div.cy_xinxi:nth-child(4) > span > a").Text()
	description := doc.Find("p#comic-description").Text()
	coverURL, ok := doc.Find("div.cy_info_cover > a > img").Attr("src")
	// The coverURL is still using 55888, manual replace it.
	coverURL = strings.Replace(coverURL, ":55888", "", 1)
	if title == "" || author == "" || description == "'" || !ok {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

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
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, *url)
		}
	})

	return &entity.Comic{
		ComicID:       *id,
		DNS:           *fn.dns,
		Title:         title,
		Author:        author,
		Description:   description,
		URL:           *url,
		Chapters:      chapters,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// UpdateComicInfo export
func (fn *FetcherNokiacn) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := fn.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

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
			logrus.Printf("Chapter missing something, title: %s, url: %s", comic.Title, comic.URL)
		}
	})

	comic.Chapters = chapters
	comic.LastCrawlTime = time.Now()
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
