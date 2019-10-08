package usecase

import (
	"log"
	"strings"

	//"strconv"
	"regexp"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
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
func (f9 *Fetcher99Comic) FetchComicInfo(url *string) *entity.Comic {
	doc := f9.FetchDoc(url)

	id := f9.GenerateID(url)
	title := doc.Find("div.comic_deCon.autoHeight > h1").Text()
	author := strings.Split(doc.Find("ul.comic_deCon_liO > li:first-child").Text(), "：")[1]
	description := strings.Replace(doc.Find("p.comic_deCon_d").Text(), " ", "", -1)
	coverURL, ok := doc.Find("div.comic_i_img > img").Attr("src")
	if title == "" || author == "" || description == "'" || !ok {
		log.Printf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
		return nil
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
			log.Printf("Chapter missing something, title: %s, url: %s", title, *url)
		}
	})

	return &entity.Comic{
		ComicID:       *id,
		DNS:           *f9.dns,
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
func (f9 *Fetcher99Comic) UpdateComicInfo(comic *entity.Comic) *entity.Comic {
	doc := f9.FetchDoc(&comic.URL)

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
			log.Printf("Chapter missing something, title: %s, url: %s", comic.Title, comic.URL)
		}
	})

	comic.Chapters = chapters
	comic.LastCrawlTime = time.Now()
	return comic
}

// FetchComicChapter export
func (f9 *Fetcher99Comic) FetchComicChapter(comic *entity.Comic, index int) []string {
	comicURLs := []string{}
	firstPageURL := f9.GetChapterURL(comic, comic.Chapters[index].URL)
	firstPage, _ := f9.FetchDoc(firstPageURL).Html()

	rChapterImages, _ := regexp.Compile(`var chapterImages = \[.*?\];`)
	rImages, _ := regexp.Compile(`".*?"`)
	images := rImages.FindAllString(rChapterImages.FindString(firstPage), -1)
	for i := 0; i < len(images); i++ {
		images[i] = images[i][1 : len(images[i])-1]
		imageURL := "http://comic.allviki.com/" + strings.Replace(images[i], `\/`, "/", 1)
		comicURLs = append(comicURLs, imageURL)
	}
	return comicURLs
}
