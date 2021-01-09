package usecase

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

// FetcherMangabz export
type FetcherMangabz struct {
	Fetcher
}

// NewFetcherMangabz export
func NewFetcherMangabz(dns string) *FetcherMangabz {
	fm := new(FetcherMangabz)
	fm.NewFetcher(false, &dns)
	return fm
}

// GetChapterURL export
func (fm *FetcherMangabz) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "http://" + comic.DNS + chapterURL
	return &url
}

// CrawlComic export
func (fm *FetcherMangabz) CrawlComic(url *string) (*entity.Comic, error) {
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
func (fm *FetcherMangabz) FetchComicInfo(comicID *string, doc *goquery.Document, cookie []*http.Cookie) (*entity.ComicInfo, error) {
	title := doc.Find("p.detail-info-title").Text()
	author := doc.Find("p.detail-info-tip > span > a").Text()
	description := doc.Find("p.detail-info-content").Text()
	coverURL, ok := doc.Find("img.detail-info-cover").Attr("src")
	if title == "" || author == "" || description == "" || !ok {
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
func (fm *FetcherMangabz) FetchChapterInfo(doc *goquery.Document, cookie []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	doc.Find("div#chapterlistload > a").Each(func(i int, s *goquery.Selection) {
		chapterTitle := s.Text()
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
func (fm *FetcherMangabz) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
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
func (fm *FetcherMangabz) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	firstPageURL := fm.GetChapterURL(comic, comic.Chapters[index].URL)
	doc, docErr := fm.FetchDoc(firstPageURL)
	if docErr != nil {
		return nil, docErr
	}
	firstPage, _ := doc.Html()
	pageExp, _ := regexp.Compile(`var MANGABZ_IMAGE_COUNT=.*?;`)
	midExp, _ := regexp.Compile(`var MANGABZ_MID=.*?;`)
	cidExp, _ := regexp.Compile(`var MANGABZ_CID=.*?;`)
	dtExp, _ := regexp.Compile(`var MANGABZ_VIEWSIGN_DT=\".*?\";`)
	signExp, _ := regexp.Compile(`var MANGABZ_VIEWSIGN=\".*?\";`)
	pageCountStr := strings.TrimRight(pageExp.FindString(firstPage)[24:], ";")
	pageCount, _ := strconv.ParseInt(pageCountStr, 10, 64)
	mid := strings.TrimRight(midExp.FindString(firstPage)[16:], ";")
	cid := strings.TrimRight(cidExp.FindString(firstPage)[16:], ";")
	dt := strings.TrimRight(dtExp.FindString(firstPage)[25:], "\";")
	sign := strings.TrimRight(signExp.FindString(firstPage)[22:], "\";")

	comicURLs := make([]string, pageCount)
	for i := int64(1); i <= pageCount; i++ {
		url := fmt.Sprintf(
			"%schapterimage.ashx?cid=%s&page=%d&key=&_cid=%s&_mid=%s&_dt=%s&_sign=%s",
			*firstPageURL,
			cid,
			i,
			cid,
			mid,
			url.QueryEscape(dt),
			sign,
		)
		plainText, _ := fm.touchImage(firstPageURL, &url)
		jsVM := otto.New()
		result, _ := jsVM.Run(*plainText)
		resultString, _ := result.ToString()
		imageURLs := strings.Split(resultString, ",")
		pageJpgExp, _ := regexp.Compile(fmt.Sprintf(`%d_.*?\.jpg`, i))
		for j := 0; j < len(imageURLs); j++ {
			if pageJpgExp.MatchString(imageURLs[j]) {
				comicURLs[i-1] = imageURLs[j]
			}
		}
	}
	return comicURLs, nil
}

func (fm *FetcherMangabz) touchImage(refererURL, url *string) (*string, error) {
	cli := &http.Client{}
	req, _ := http.NewRequest("GET", *url, nil)
	req.Header.Set("Referer", *refererURL)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	res, err := cli.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "When TouchImage")
	}
	data, _ := ioutil.ReadAll(res.Body)
	stringData := string(data)

	return &stringData, nil
}
