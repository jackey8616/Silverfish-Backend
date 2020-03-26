package usecase

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	//"strconv"

	"time"

	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
)

// FetcherComicbus export
type FetcherComicbus struct {
	Fetcher
}

// NewFetcherComicbus export
func NewFetcherComicbus(dns string) *FetcherComicbus {
	fc := new(FetcherComicbus)
	fc.NewFetcher(true, &dns)
	return fc
}

// GetChapterURL export
func (fc *FetcherComicbus) GetChapterURL(comic *entity.Comic, chapterURL string) *string {
	url := "https://" + comic.DNS + chapterURL
	return &url
}

// FetchComicInfo export
func (fc *FetcherComicbus) FetchComicInfo(url *string) (*entity.Comic, error) {
	doc, cookies, docErr := fc.FetchDocWithEncoding(url, "Big5")
	if docErr != nil {
		return nil, docErr
	}

	id := fc.GenerateID(url)
	title := doc.Find("td > font[style='font-size:10pt; letter-spacing:1px']").Text()
	author := doc.Find("td:contains('作者：')+td").Text()
	description := doc.Find("tbody > tr > td[style='line-height:25px']").Text()
	coverURL, ok := doc.Find("img[style*='border:#CCCCCC solid 1px;']").Attr("src")
	coverURL = "https://comicbus.com/" + coverURL
	if title == "" || author == "" || description == "" || !ok {
		return nil, fmt.Errorf("Something missing, title: %s, author: %s, description: %s, coverURL: %s", title, author, description, coverURL)
	}

	chapters := []entity.ComicChapter{}
	doc.Find("table#rp_ctl05_0_dl_0 > tbody > tr").Each(func(i int, s *goquery.Selection) {
		s.Find("td[style='width:10%;white-space:nowrap;'] > a").Each(func(j int, r *goquery.Selection) {
			cview, ok := r.Attr("onclick")
			if ok {
				click := strings.Split(cview, ",")
				click[0] = click[0][strings.Index(click[0], "'")+1 : strings.LastIndex(click[0], "'")]
				click[2] = click[2][0:strings.Index(click[2], ")")]
				chapterTitle := strings.Replace(r.Text(), "\n", "", 1)
				chapterURL := *fc.cview(&click[0], &click[1], &click[2], cookies[0])
				chapters = append(chapters, entity.ComicChapter{
					Title:    chapterTitle,
					URL:      chapterURL,
					ImageURL: []string{},
				})
			} else {
				log.Printf("Chapter missing something, title: %s, url: %s", title, *url)
			}
		})
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
	}, nil
}

// UpdateComicInfo export
func (fc *FetcherComicbus) UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error) {
	doc, cookies, docErr := fc.FetchDocWithEncoding(&comic.URL, "Big5")
	if docErr != nil {
		return nil, docErr
	}

	chapters := []entity.ComicChapter{}
	doc.Find("table#rp_ctl05_0_dl_0 > tbody > tr").Each(func(i int, s *goquery.Selection) {
		s.Find("td[style='width:10%;white-space:nowrap;'] > a").Each(func(j int, r *goquery.Selection) {
			cview, ok := r.Attr("onclick")
			if ok {
				click := strings.Split(cview, ",")
				click[0] = click[0][strings.Index(click[0], "'")+1 : strings.LastIndex(click[0], "'")]
				click[2] = click[2][0:strings.Index(click[2], ")")]
				chapterTitle := strings.Replace(r.Text(), "\n", "", 1)
				chapterURL := *fc.cview(&click[0], &click[1], &click[2], cookies[0])
				chapters = append(chapters, entity.ComicChapter{
					Title:    chapterTitle,
					URL:      chapterURL,
					ImageURL: []string{},
				})
			} else {
				log.Printf("Chapter missing something, title: %s, url: %s", comic.Title, comic.URL)
			}
		})
	})

	comic.Chapters = chapters
	comic.LastCrawlTime = time.Now()
	return comic, nil
}

// FetchComicChapter export
func (fc *FetcherComicbus) FetchComicChapter(comic *entity.Comic, index int) ([]string, error) {
	comicURLs := []string{}
	firstPageURL := comic.Chapters[index].URL
	doc, docErr := fc.FetchDoc(&firstPageURL)
	if docErr != nil {
		return nil, docErr
	}
	form, _ := doc.Find("form#Form1 > script").Html()
	firstURL, pageCount := fc.initPage(index+1, 1, &form)
	comicURLs = append(comicURLs, *firstURL)

	for i := 1; i < pageCount; i++ {
		imageURL, _ := fc.initPage(index+1, i+1, &form)
		comicURLs = append(comicURLs, *imageURL)
	}
	return comicURLs, nil
}

func (fc *FetcherComicbus) cview(url, catid, copyright *string, cookie *http.Cookie) *string {
	inputURL := strings.Replace(*url, ".html", "", 1)
	inputURL = strings.Replace(inputURL, "-", ".html?ch=", 1)

	RI := &cookie.Value
	baseURL := "https://www.comicbus.xyz"
	if *RI == "3" && *copyright == "1" {
		baseURL = "https://www.comicbus.xyz/online/c-"
	} else {
		if *copyright == "1" {
			switch *catid {
			case "4", "6", "12", "22":
				baseURL = "https://www.comicbus.xyz/online/comic-"
			case "1", "17", "19", "21":
				baseURL = "https://www.comicbus.xyz/online/comic-"
			case "2", "5", "7", "9":
				baseURL = "https://www.comicbus.xyz/online/comic-"
			case "10", "11", "13", "14":
				baseURL = "https://www.comicbus.xyz/online/comic-"
			case "3", "8", "15", "16", "18", "20":
				baseURL = "https://www.comicbus.xyz/online/comic-"
			}
		} else {
			baseURL = "https://www.comicbus.xyz/online/manga_"
		}
	}
	chapterURL := baseURL + inputURL
	return &chapterURL
}

func (fc *FetcherComicbus) initPage(chapter, page int, script *string) (*string, int) {
	normalScript := strings.Replace(*script, "&#39;", "'", -1)
	normalScript = strings.Replace(normalScript, "&#34;", "\"", -1)
	normalScript = strings.Replace(normalScript, "&gt;", ">", -1)
	normalScript = strings.Replace(normalScript, "&lt;", "<", -1)

	normalScript = normalScript[strings.Index(normalScript, "ch=parseInt(ch);")+16 : len(normalScript)-6]
	normalScript = strings.Replace(normalScript, normalScript[strings.Index(normalScript, "function"):strings.Index(normalScript, "}")+1], "", 1)
	normalScript = fmt.Sprintf("var url='';var ch=%d;var p=%d;%s", chapter, page, normalScript)
	normalScript = strings.Replace(normalScript, "ge('TheImg').src='//img'", "url = 'https://img'", 1)

	functionScript := `var y = 46;
	function lc(l) {
		if (l.length != 2) return l;
		var az = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";
		var a = l.substring(0, 1);
		var b = l.substring(1, 2);
		if (a == "Z") return 8000 + az.indexOf(b);
		else return az.indexOf(a) * 52 + az.indexOf(b);
	}
	function su(a, b, c) {
		var e = (a + '').substring(b, b + c);
		return (e);
	}
	function nn(n) {
		return n < 10 ? '00' + n : n < 100 ? '0' + n : n;
	}

	function mm(p) {
		return (parseInt((p - 1) / 10) % 10) + (((p - 1) % 10) * 3)
	};`

	executeScript := functionScript + normalScript
	jsVM := otto.New()
	jsVM.Run(executeScript)
	psVal, _ := jsVM.Get("ps")
	urlVal, _ := jsVM.Get("url")

	pages, _ := psVal.ToInteger()
	URL, _ := urlVal.ToString()
	return &URL, int(pages)
}
