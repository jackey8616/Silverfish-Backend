package silverfish

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"gopkg.in/mgo.v2/bson"
)

// Silverfish export
type Silverfish struct {
	mgoInf *MongoInf
}

// New export
func New(mgoInf *MongoInf) *Silverfish {
	sf := new(Silverfish)
	sf.mgoInf = mgoInf
	return sf
}

// FetchNovel export
func (sf *Silverfish) FetchNovel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	data := map[string]string{}
	decoder.Decode(&data)

	targetURL := data["novel_url"]
	result, err := sf.mgoInf.FindOne(bson.M{"url": targetURL}, &Novel{})
	record, _ := result.(*Novel)

	if err != nil || time.Now().Sub(record.LastCrawlTime) > 24*time.Hour {
		log.Println("Missing or expired record, recrawl.")
		doc, er := goquery.NewDocument(targetURL)
		if er != nil {
			log.Fatal(er)
		}

		charset, _ := doc.Find("meta[content*='charset']").Attr("content")
		charset = charset[strings.LastIndex(charset, "=")+1:]
		title, _ := doc.Find("meta[property='og:title']").Attr("content")
		coverURL, _ := doc.Find("meta[property='og:image']").Attr("content")
		enc := mahonia.NewDecoder(charset)

		chapters := []Chapter{}
		doc.Find("div#list-chapterAll > dl > dd > a").Each(func(i int, s *goquery.Selection) {
			chapterTitle, _ := s.Attr("title")
			chapterURL, _ := s.Attr("href")
			chapters = append(chapters, Chapter{
				Title: enc.ConvertString(chapterTitle),
				URL:   chapterURL,
			})
		})

		record = &Novel{
			Title:         enc.ConvertString(title),
			URL:           targetURL,
			CoverURL:      coverURL,
			Chapters:      chapters,
			LastCrawlTime: time.Now(),
		}
		sf.mgoInf.Upsert(bson.M{"url": targetURL}, record)
	}

	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]Novel{"Rtn": *record})
	w.Write(js)
}

// FetchChapter export
func (sf *Silverfish) FetchChapter(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	data := map[string]string{}
	decoder.Decode(&data)

	targetURL := data["chapter_url"]
	doc, err := goquery.NewDocument(targetURL)
	if err != nil {
		log.Fatal(err)
	}

	html, _ := doc.Html()
	charset, _ := doc.Find("meta[content*='charset']").Attr("content")
	enc := mahonia.NewDecoder(charset)
	rawHTML := enc.ConvertString(html)

	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]string{"Rtn": rawHTML})
	w.Write(js)
}
