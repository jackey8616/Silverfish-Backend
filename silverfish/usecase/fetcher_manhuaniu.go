package usecase

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	//"strconv"

	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// FetcherManhuaniu export
type FetcherManhuaniu struct {
	Fetcher
}

// NewFetcherManhuaniu export
func NewFetcherManhuaniu(dns string) *FetcherManhuaniu {
	fm := new(FetcherManhuaniu)
	fm.NewFetcher(true, &dns)
	return fm
}

// GetChapterURL export
func (fm *FetcherManhuaniu) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "https://" + comic.DNS + chapterURL
	return &url
}

// CrawlComic export
func (fm *FetcherManhuaniu) CrawlComic(url *string) (*entity.Comic, error) {
	doc, docErr := fm.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fm.GenerateId(url)
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
func (fm *FetcherManhuaniu) FetchComicInfo(comicId *string, doc *goquery.Document, cookies []*http.Cookie) (*entity.ComicInfo, error) {
	title := doc.Find("div.book-title > h1 > span").Text()
	author := doc.Find("ul.detail-list > li:nth-child(2) > span:nth-child(2) > a").Text()
	description := doc.Find("div#intro-all > p").Text()
	coverURL, ok := doc.Find("div.book-cover > p.cover > img").Attr("src")
	if title == "" || author == "" || description == "'" || !ok {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	return &entity.ComicInfo{
		IsEnable:      true,
		ComicId:       *comicId,
		Title:         title,
		Author:        author,
		Description:   description,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fm *FetcherManhuaniu) FetchChapterInfo(doc *goquery.Document, cookies []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	doc.Find("div.chapter-body > ul > li > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Find("span").Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapters = append(chapters, entity.ComicChapter{
				Title:    chapterTitle,
				URL:      chapterURL,
				ImageURL: []string{},
			})
		} else {
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})

	return chapters
}

// UpdateComicInfo export
func (fm *FetcherManhuaniu) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := fm.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fm.FetchComicInfo(&comic.ComicId, doc, nil)
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
func (fm *FetcherManhuaniu) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	firstPageURL := fm.GetChapterURL(comic, comic.Chapters[index].URL)
	doc, docErr := fm.FetchDoc(firstPageURL)
	if docErr != nil {
		return nil, docErr
	}
	firstPage, _ := doc.Html()

	rChapterImages, _ := regexp.Compile(`var chapterImages = \[.*?\];`)
	rImages, _ := regexp.Compile(`".*?"`)
	images := rImages.FindAllString(rChapterImages.FindString(firstPage), -1)
	for i := 0; i < len(images); i++ {
		images[i] = images[i][1 : len(images[i])-1]
		imageURL := "https://res.nbhbzl.com/" + strings.Replace(images[i], `\/`, "/", -1)
		comicURLs = append(comicURLs, imageURL)
	}
	return comicURLs, nil
}
