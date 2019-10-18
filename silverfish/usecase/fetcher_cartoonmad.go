package usecase

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// FetcherCartoonmad export
type FetcherCartoonmad struct {
	Fetcher
}

// NewFetcherCartoonmad export
func NewFetcherCartoonmad(dns string) *FetcherCartoonmad {
	fc := new(FetcherCartoonmad)
	fc.NewFetcher(true, &dns)
	return fc
}

// GetChapterURL export
func (fc *FetcherCartoonmad) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "https://" + comic.DNS + chapterURL
	return &url
}

// FetchComicInfo export
func (fc *FetcherCartoonmad) FetchComicInfo(url *string) *entity.Comic {
	if strings.Contains(*url, "/m/") == false {
		temp := strings.Replace(*url, "/comic/", "/m/comic/", 1)
		url = &temp
	}
	doc, _ := fc.FetchDocWithEncoding(url, "Big5")
	id := fc.GenerateID(url)

	anchor := doc.Find("table:nth-of-type(2) > tbody > tr:nth-of-type(2) > td")
	title, err0 := doc.Find("meta[name='Keywords']").Attr("content")
	title = strings.Split(title, ",")[0]
	author := anchor.Find("table > tbody > tr:nth-of-type(4) > td").Text()
	author = strings.Split(author, "作者：")[1]
	author = strings.Split(author, "‧")[0]

	description := anchor.Find("table:nth-of-type(2) > tbody > tr > td > fieldset > table > tbody > tr > td").Text()
	coverURL, err1 := anchor.Find("table > tbody > tr > td:nth-of-type(2) > img").Attr("src")
	coverURL = "https://" + *fc.dns + coverURL
	if !err0 || author == "" || description == "" || !err1 {
		log.Printf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
		return nil
	}

	chapters := []entity.ComicChapter{}
	anchor.Find("table:nth-of-type(3) > tbody > tr:nth-of-type(2) > td > fieldset > table > tbody > tr[align='center'] > td > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapterURL = strings.Replace(chapterURL, "/m/comic/", "/comic/", 1)
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
		DNS:           *fc.dns,
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
func (fc *FetcherCartoonmad) UpdateComicInfo(comic *entity.Comic) *entity.Comic {
	doc, _ := fc.FetchDocWithEncoding(&comic.URL, "Big5")
	anchor := doc.Find("table:nth-of-type(2) > tbody > tr:nth-of-type(2) > td")

	chapters := []entity.ComicChapter{}
	anchor.Find("table:nth-of-type(3) > tbody > tr:nth-of-type(2) > td > fieldset > table > tbody > tr[align='center'] > td > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
		chapterURL, ok := s.Attr("href")
		if ok {
			chapterURL = strings.Replace(chapterURL, "/m/comic/", "/comic/", 1)
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
func (fc *FetcherCartoonmad) FetchComicChapter(comic *entity.Comic, index int) []string {
	comicURLs := []string{}
	url := fc.GetChapterURL(comic, comic.Chapters[index].URL)
	doc := fc.FetchDoc(url)

	lastIndex, _ := strconv.Atoi(doc.Find("table > tbody > tr:nth-of-type(6) > td > a:nth-last-of-type(2)").Text())
	partImageURL, _ := doc.Find("table > tbody > tr:nth-of-type(5) > td > table > tbody > tr > td > a > img").Attr("src")
	imageURL := fc.GetChapterURL(comic, "/comic/"+partImageURL)
	touchedURL := fc.touchImage(url, imageURL)

	for i := 1; i <= lastIndex; i++ {
		comicURLs = append(comicURLs, fmt.Sprintf("%s/%03d.jpg", touchedURL, i))
	}

	return comicURLs
}

func (fc *FetcherCartoonmad) touchImage(refererURL, url *string) string {
	cli := &http.Client{}
	req, _ := http.NewRequest("GET", *url, nil)
	req.Header.Set("Referer", *refererURL)

	res, err := cli.Do(req)
	if err != nil {
		log.Fatal(errors.Wrap(err, "When TouchImage"))
		return ""
	}

	temp := res.Request.URL.String()
	return temp[:strings.LastIndex(temp, "/")]
}
