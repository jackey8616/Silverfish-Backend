package usecase

import (
	"fmt"
	"strings"

	//"strconv"
	"regexp"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// Fetcher99Comic export
type Fetcher99Comic struct {
	Fetcher
}

// NewFetcher99Comic export
func NewFetcher99Comic(dns string) *Fetcher99Comic {
	f9 := new(Fetcher99Comic)
	f9.NewFetcher(true, &dns)
	return f9
}

// GetChapterURL export
func (f9 *Fetcher99Comic) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "https://" + comic.DNS + chapterURL
	return &url
}

// FetchComicInfo export
func (f9 *Fetcher99Comic) FetchComicInfo(url *string) (*entity.Comic, error) {
	doc, docErr := f9.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := f9.GenerateId(url)
	title := doc.Find("div.comic_deCon.autoHeight > h1").Text()
	author := strings.Split(doc.Find("ul.comic_deCon_liO > li:first-child").Text(), "：")[1]
	description := strings.Replace(doc.Find("p.comic_deCon_d").Text(), " ", "", -1)
	coverURL, ok := doc.Find("div.comic_i_img > img").Attr("src")
	if title == "" || author == "" || description == "'" || !ok {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	chapters := []entity.ComicChapter{}
	doc.Find("ul#chapter-list-3 > li > a, ul#chapter-list-2 > li > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := strings.Replace(s.Find("span.list_con_zj").Text(), " ", "", -1)
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append(chapters, entity.ComicChapter{
				Title:    chapterTitle,
				URL:      chapterURL,
				ImageURL: []string{},
			})
		} else {
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, *url)
		}
	})

	return &entity.Comic{
		ComicId:       *id,
		DNS:           *f9.dns,
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
func (f9 *Fetcher99Comic) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := f9.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

	chapters := []entity.ComicChapter{}
	doc.Find("ul#chapter-list-3 > li > a, ul#chapter-list-2 > li > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := strings.Replace(s.Find("span.list_con_zj").Text(), " ", "", -1)
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append(chapters, entity.ComicChapter{
				Title:    chapterTitle,
				URL:      chapterURL,
				ImageURL: []string{},
			})
		} else {
			logrus.Printf("Chapter missing something, title: %s, url: %s", comic.Title, comic.URL)
		}
	})

	comic.Chapters = chapters
	comic.LastCrawlTime = time.Now()
	return comic, nil
}

// FetchComicChapter export
func (f9 *Fetcher99Comic) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	firstPageURL := f9.GetChapterURL(comic, comic.Chapters[index].URL)
	doc, docErr := f9.FetchDoc(firstPageURL)
	if docErr != nil {
		return nil, docErr
	}
	firstPage, _ := doc.Html()

	rChapterImages, _ := regexp.Compile(`var chapterImages = \[.*?\];`)
	rImages, _ := regexp.Compile(`".*?"`)
	images := rImages.FindAllString(rChapterImages.FindString(firstPage), -1)
	for i := 0; i < len(images); i++ {
		images[i] = images[i][1 : len(images[i])-1]
		imageURL := "http://comic.allviki.com/" + strings.Replace(images[i], `\/`, "/", 1)
		comicURLs = append(comicURLs, imageURL)
	}
	return comicURLs, nil
}
