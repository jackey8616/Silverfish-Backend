package usecase

import (
	"fmt"
	"net/http"
	entity "silverfish/silverfish/entity"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// FetcherBaozimh export
type FetcherBaozimh struct {
	Fetcher
}

// NetFetcherBaozimh export
func NewFetcherBaozimh(dns string) *FetcherBaozimh {
	fb := new(FetcherBaozimh)
	fb.NewFetcher(true, &dns)
	return fb
}

// GetChapterURL export
func (fb *FetcherBaozimh) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "http://" + comic.DNS + chapterURL
	return &url
}

// CrawlComic export
func (fb *FetcherBaozimh) CrawlComic(url *string) (*entity.Comic, error) {
	doc, docErr := fb.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fb.GenerateID(url)
	info, infoErr := fb.FetchComicInfo(id, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fb.FetchChapterInfo(doc, nil, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic := &entity.Comic{
		DNS:      *fb.dns,
		URL:      *url,
		Chapters: chapters,
	}
	comic.SetComicInfo(info)
	return comic, nil
}

// FetchComicInfo export
func (fb *FetcherBaozimh) FetchComicInfo(comicID *string, doc *goquery.Document, cookie []*http.Cookie) (*entity.ComicInfo, error) {
	title, titleOk := doc.Find("meta[name='og:novel:book_name']").Attr("content")
	author, authorOk := doc.Find("meta[name='og:novel:author']").Attr("content")
	description := strings.TrimSpace(strings.Replace(doc.Find("p.comics-detail__desc.overflow-hidden").Text(), `\n`, ``, -1))
	coverURL, coverOk := doc.Find("meta[name='og:image']").Attr("content")
	if !titleOk || !authorOk || description == "" || !coverOk {
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
func (fb *FetcherBaozimh) FetchChapterInfo(doc *goquery.Document, cookie []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	doc.Find("#chapter-items > div > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Find("div > span").Text()
		chapterUrl, ok := s.Attr("href")
		if chapterTitle != "" && ok {
			chapters = append(chapters, entity.ComicChapter{
				Title:    chapterTitle,
				URL:      chapterUrl,
				ImageURL: []string{},
			})
		} else {
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})
	doc.Find("#chapters_other_list > div > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Find("div > span").Text()
		chapterUrl, ok := s.Attr("href")
		if chapterTitle != "" && ok {
			chapters = append(chapters, entity.ComicChapter{
				Title:    chapterTitle,
				URL:      chapterUrl,
				ImageURL: []string{},
			})
		} else {
			logrus.Printf("Chapter missing something, title: %s, url: %s", title, url)
		}
	})
	return chapters
}

// UpdateComicInfo export
func (fb *FetcherBaozimh) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := fb.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fb.FetchComicInfo(&comic.ComicID, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fb.FetchChapterInfo(doc, nil, comic.Title, comic.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic.LastCrawlTime = time.Now()
	comic.SetComicInfo(info)
	comic.Chapters = chapters
	return comic, nil
}

// FetchComicChapter export
func (fb *FetcherBaozimh) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	url := fb.GetChapterURL(comic, comic.Chapters[index].URL)
	doc, docErr := fb.FetchDoc(url)
	if docErr != nil {
		return nil, fmt.Errorf("FetchComicChapter error")
	}

	doc.Find("ul.comic-contain > div > amp-img").Each(func(i int, s *goquery.Selection) {
		imageUrl, ok := s.Attr("src")
		if imageUrl != "" && ok {
			comicURLs = append(comicURLs, imageUrl)
		} else {
			logrus.Printf("ChapterImage missing url")
		}
	})
	logrus.Print(comicURLs)
	return comicURLs, nil
}
