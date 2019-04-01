package silverfish

import (
	"encoding/json"
	"fmt"
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

// Proxy export
func (sf *Silverfish) Proxy(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	data := map[string]string{}
	decoder.Decode(&data)

	targetURL := data["proxy_url"]
	result, err := sf.mgoInf.FindOne(bson.M{"url": targetURL}, &Novel{})
	record, _ := result.(*Novel)

	if err != nil || time.Now().Sub(record.LastCrawlTime) > 24*time.Hour {
		log.Println("Missing or expired record, recrawl.")
		doc, er := goquery.NewDocument(targetURL)
		if er != nil {
			fmt.Println(er)
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
