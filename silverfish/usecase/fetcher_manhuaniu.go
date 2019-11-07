package usecase

import (
	"log"
	"regexp"
	"strings"

	//"strconv"

	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
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

// FetchComicInfo export
func (fm *FetcherManhuaniu) FetchComicInfo(url *string) *entity.Comic {
	doc := fm.FetchDoc(url)

	id := fm.GenerateID(url)
	title := doc.Find("div.book-title > h1 > span").Text()
	author := doc.Find("ul.detail-list > li:nth-child(2) > span:nth-child(2) > a").Text()
	description := doc.Find("div#intro-all > p").Text()
	coverURL, ok := doc.Find("div.book-cover > p.cover > img").Attr("src")
	if title == "" || author == "" || description == "" || !ok {
		log.Printf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
		return nil
	}

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
			log.Printf("Chapter missing something, title: %s, url: %s", title, *url)
		}
	})

	return &entity.Comic{
		ComicID:       *id,
		DNS:           *fm.dns,
		Title:         title,
		Author:        author,
		Description:   description,
		URL:           *url,
		Chapters:      chapters,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}
}

// UpdateComicInfo export
func (fm *FetcherManhuaniu) UpdateComicInfo(comic *entity.Comic) *entity.Comic {
	doc := fm.FetchDoc(&comic.URL)

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
			log.Printf("Chapter missing something, title: %s, url: %s", comic.Title, comic.URL)
		}
	})

	comic.Chapters = chapters
	comic.LastCrawlTime = time.Now()
	return comic
}

// FetchComicChapter export
func (fm *FetcherManhuaniu) FetchComicChapter(comic *entity.Comic, index int) []string {
	comicURLs := []string{}
	firstPageURL := fm.GetChapterURL(comic, comic.Chapters[index].URL)
	firstPage, _ := fm.FetchDoc(firstPageURL).Html()

	rChapterImages, _ := regexp.Compile(`var chapterImages = \[.*?\];`)
	rImages, _ := regexp.Compile(`".*?"`)
	images := rImages.FindAllString(rChapterImages.FindString(firstPage), -1)
	for i := 0; i < len(images); i++ {
		images[i] = images[i][1 : len(images[i])-1]
		imageURL := "https://res.nbhbzl.com/" + strings.Replace(images[i], `\/`, "/", -1)
		comicURLs = append(comicURLs, imageURL)
	}
	return comicURLs
}
