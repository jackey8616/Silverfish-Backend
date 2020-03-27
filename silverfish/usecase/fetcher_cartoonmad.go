package usecase

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

// CrawlComic export
func (fc *FetcherCartoonmad) CrawlComic(url *string) (*entity.Comic, error) {
	if strings.Contains(*url, "/m/") == false {
		temp := strings.Replace(*url, "/comic/", "/m/comic/", 1)
		url = &temp
	}
	doc, _, docErr := fc.FetchDocWithEncoding(url, "Big5")
	if docErr != nil {
		return nil, docErr
	}
	id := fc.GenerateID(url)

	info, infoErr := fc.FetchComicInfo(id, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fc.FetchChapterInfo(doc, nil, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic := &entity.Comic{
		DNS:      *fc.dns,
		URL:      *url,
		Chapters: chapters,
	}
	comic.SetComicInfo(info)
	return comic, nil
}

// FetchComicInfo export
func (fc *FetcherCartoonmad) FetchComicInfo(comicID *string, doc *goquery.Document, cookies []*http.Cookie) (*entity.ComicInfo, error) {
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
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	return &entity.ComicInfo{
		ComicID:       *comicID,
		Title:         title,
		Author:        author,
		Description:   description,
		CoverURL:      coverURL,
		LastCrawlTime: time.Now(),
	}, nil
}

// FetchChapterInfo export
func (fc *FetcherCartoonmad) FetchChapterInfo(doc *goquery.Document, cookies []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	doc.Find("table:nth-of-type(2) > tbody > tr:nth-of-type(2) > td > table:nth-of-type(3) > tbody > tr:nth-of-type(2) > td > fieldset > table > tbody > tr[align='center'] > td > a").Each(func(i int, s *goquery.Selection) {
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
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})

	return chapters
}

// UpdateComicInfo export
func (fc *FetcherCartoonmad) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, _, docErr := fc.FetchDocWithEncoding(&comic.URL, "Big5")
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fc.FetchComicInfo(&comic.ComicID, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fc.FetchChapterInfo(doc, nil, info.Title, comic.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic.LastCrawlTime = time.Now()
	comic.SetComicInfo(info)
	comic.Chapters = chapters
	return comic, nil
}

// FetchComicChapter export
func (fc *FetcherCartoonmad) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	url := fc.GetChapterURL(comic, comic.Chapters[index].URL)
	doc, docErr := fc.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	lastIndex, _ := strconv.Atoi(doc.Find("table > tbody > tr:nth-of-type(6) > td > a:nth-last-of-type(2)").Text())
	partImageURL, _ := doc.Find("table > tbody > tr:nth-of-type(5) > td > table > tbody > tr > td > a > img").Attr("src")
	imageURL := fc.GetChapterURL(comic, "/comic/"+partImageURL)
	touchedURL := fc.touchImage(url, imageURL)

	for i := 1; i <= lastIndex; i++ {
		comicURLs = append(comicURLs, fmt.Sprintf("%s/%03d.jpg", touchedURL, i))
	}

	return comicURLs, nil
}

func (fc *FetcherCartoonmad) touchImage(refererURL, url *string) string {
	cli := &http.Client{}
	req, _ := http.NewRequest("GET", *url, nil)
	req.Header.Set("Referer", *refererURL)

	res, err := cli.Do(req)
	if err != nil {
		logrus.Print(errors.Wrap(err, "When TouchImage"))
		return ""
	}

	temp := res.Request.URL.String()
	return temp[:strings.LastIndex(temp, "/")]
}
