package silverfish

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackey8616/Silverfish-backend/silverfish/entity"
	"github.com/jackey8616/Silverfish-backend/silverfish/usecase"
	"gopkg.in/mgo.v2/bson"
)

// Silverfish export
type Silverfish struct {
	mgoInf   *entity.MongoInf
	fetchers []entity.NovelFetcher
}

// New export
func New(mgoInf *entity.MongoInf) *Silverfish {
	sf := new(Silverfish)
	sf.mgoInf = mgoInf
	sf.fetchers = []entity.NovelFetcher{
		usecase.NewFetcher77xsw(),
	}
	return sf
}

// FetchNovel export
func (sf *Silverfish) FetchNovel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	data := map[string]string{}
	decoder.Decode(&data)

	targetURL := data["novel_url"]
	result, err := sf.mgoInf.FindOne(bson.M{"url": targetURL}, &entity.Novel{})
	record := result.(*entity.Novel)

	// Err or need to recrawl
	if err != nil || time.Now().Sub(record.LastCrawlTime) > 24*time.Hour {
		log.Println("Missing or expired record, recrawl.")
		for _, each := range sf.fetchers {
			if each.Match(&targetURL) {
				record = each.FetchNovelInfo(&targetURL)
				sf.mgoInf.Upsert(bson.M{"novelID": record.NovelID}, record)
				break
			}
			log.Fatal(fmt.Sprintf("No suit crawler for %s", targetURL))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]entity.Novel{"Rtn": *record})
	w.Write(js)
}

// FetchChapter export
func (sf *Silverfish) FetchChapter(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	data := map[string]string{}
	decoder.Decode(&data)

	targetURL := data["chapter_url"]
	rawHTML := ""
	for _, each := range sf.fetchers {
		if each.Match(&targetURL) {
			rawHTML = *each.FetchChapter(&targetURL)
			break
		}
		log.Fatal(fmt.Sprintf("No suit crawler for %s", targetURL))
	}

	w.Header().Set("Content-Type", "application/json")
	js, _ := json.Marshal(map[string]string{"Rtn": rawHTML})
	w.Write(js)
}
