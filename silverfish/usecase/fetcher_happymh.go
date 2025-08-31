package usecase

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// FetcherHappymh export
type FetcherHappymh struct {
	Fetcher
}

// NewFetcherHappymh export
func NewFetcherHappymh(dns string) *FetcherHappymh {
	fh := new(FetcherHappymh)
	fh.NewFetcher(false, &dns)
	return fh
}

// GetChapterURL export
func (fh *FetcherHappymh) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "https://" + comic.DNS + chapterURL
	return &url
}

// CrawlComic export
func (fh *FetcherHappymh) CrawlComic(url *string) (*entity.Comic, error) {
	doc, docErr := fh.FetchDoc(url)
	if docErr != nil {
		return nil, docErr
	}

	id := fh.GenerateID(url)
	info, infoErr := fh.FetchComicInfo(id, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fh.FetchChapterInfo(doc, nil, info.Title, *url)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic := &entity.Comic{
		DNS:      *fh.dns,
		URL:      *url,
		Chapters: chapters,
	}
	comic.SetComicInfo(info)
	return comic, nil
}

// FetchComicInfo export
func (fh *FetcherHappymh) FetchComicInfo(comicID *string, doc *goquery.Document, cookie []*http.Cookie) (*entity.ComicInfo, error) {
	title := doc.Find("h2.mg-title").Text()
	author := doc.Find("p.mg-sub-title:nth-of-type(2) > a").Text()
	description := doc.Find("div.manga-introduction > mip-showmore").Text()
	coverURL, ok := doc.Find("div.mg-cover > mip-img").Attr("src")
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
func (fh *FetcherHappymh) FetchChapterInfo(doc *goquery.Document, cookie []*http.Cookie, title, url string) []entity.ComicChapter {
	chapters := []entity.ComicChapter{}
	textData := doc.Find("mip-data:nth-of-type(3)").Text()
	exp, _ := regexp.Compile(`"chapterList":\[.*?\],`)
	data := exp.FindString(textData)
	arrayDataStr, _ := strconv.Unquote(strings.Replace(strconv.Quote(data[16:len(data)-2]), `\\u`, `\u`, -1))
	chapterDatas := strings.Split(arrayDataStr, "},{")

	titleExp, _ := regexp.Compile(`\"chapterName\":\".*?\"`)
	idExp, _ := regexp.Compile(`\"id\":\".*?\"`)
	for i := 0; i < len(chapterDatas); i++ {
		title := titleExp.FindString(chapterDatas[i])[15:]
		id := idExp.FindString(chapterDatas[i])[6:]

		chapters = append([]entity.ComicChapter{
			entity.ComicChapter{
				Title:    title[:len(title)-1],
				URL:      fmt.Sprintf("%s/%s", url, id[:len(id)-1]),
				ImageURL: []string{},
			},
		}, chapters...)
	}
	return chapters
}

// UpdateComicInfo export
func (fh *FetcherHappymh) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, docErr := fh.FetchDoc(&comic.URL)
	if docErr != nil {
		return nil, docErr
	}

	info, infoErr := fh.FetchComicInfo(&comic.ComicID, doc, nil)
	if infoErr != nil {
		return nil, fmt.Errorf("Something wrong while fetching info: %s", infoErr.Error())
	}
	chapters := fh.FetchChapterInfo(doc, nil, comic.Title, comic.URL)
	if len(chapters) == 0 {
		logrus.Print("Chapters is empty. Strange...")
	}

	comic.LastCrawlTime = time.Now()
	comic.SetComicInfo(info)
	comic.Chapters = chapters
	return comic, nil
}

// FetchComicChapter export
func (fh *FetcherHappymh) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	browser := fh.GenerateRodBrowser()
	defer browser.MustClose()

	page := browser.MustConnect().MustPage(comic.Chapters[index].URL)
	page.Race().Element("div#iframeContainer_0").MustHandle(func(e *rod.Element) {
		imgJSONs := page.MustEval("JSON.parse(ReadJs.dct('c1zbnttrabim', ss))").Arr()
		for i := 0; i < len(imgJSONs); i++ {
			comicURLs = append(comicURLs, imgJSONs[i].Get("url").String())
		}
	}).MustDo()
	return comicURLs, nil
}
