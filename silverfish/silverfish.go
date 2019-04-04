package silverfish

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackey8616/Silverfish-backend/silverfish/entity"
	"github.com/jackey8616/Silverfish-backend/silverfish/usecase"
	"gopkg.in/mgo.v2/bson"
)

// Silverfish export
type Silverfish struct {
	Router   *Router
	mgoInf   *entity.MongoInf
	fetchers map[string]entity.NovelFetcher
	urls     []string
}

// New export
func New(mgoInf *entity.MongoInf, urls []string) *Silverfish {
	sf := new(Silverfish)
	sf.Router = NewRouter(sf)
	sf.mgoInf = mgoInf
	sf.urls = urls
	sf.fetchers = map[string]entity.NovelFetcher{
		"www.77xsw.la": usecase.NewFetcher77xsw("www.77xsw.la"),
	}
	return sf
}

// getNovels
func (sf *Silverfish) getNovels() *entity.APIResponse {
	result, err := sf.mgoInf.FindSelectAll(nil, bson.M{
		"novelID": 1, "coverUrl": 1, "title": 1, "lastCrawlTime": 1}, &[]entity.NovelInfo{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	return &entity.APIResponse{
		Success: true,
		Data:    result.(*[]entity.NovelInfo),
	}
}

// getNovel
func (sf *Silverfish) getNovel(novelID *string) *entity.APIResponse {
	result, err := sf.mgoInf.FindOne(bson.M{"novelID": *novelID}, &entity.Novel{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	return &entity.APIResponse{
		Success: true,
		Data:    result.(*entity.Novel),
	}
}

// getChapter
func (sf *Silverfish) getChapter(novelID *string) *entity.APIResponse {
	query, err := sf.mgoInf.FindOne(bson.M{"novelID": novelID}, &entity.Novel{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	record := query.(*entity.Novel)
	if val, ok := sf.fetchers[record.DNS]; ok {
		return &entity.APIResponse{
			Success: true,
			Data:    val.FetchChapter(&record.NovelID),
		}
	}
	return &entity.APIResponse{
		Fail: true,
		Data: map[string]string{"reason": "No such fetcher."},
	}
}

func (sf *Silverfish) getChapterNew(novelID, chapterIndex *string) *entity.APIResponse {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": "Invalid chapter index"},
		}
	}
	query, err := sf.mgoInf.FindOne(bson.M{"novelID": novelID}, &entity.Novel{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	record := query.(*entity.Novel)
	if len((*record).Chapters) < index {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": "Wrong Index"},
		}
	} else if val, ok := sf.fetchers[*novelID]; ok {
		return &entity.APIResponse{
			Success: true,
			Data:    val.FetcherNewChapter(record, index),
		}
	}
	return &entity.APIResponse{
		Success: true,
		Data:    map[string]string{"reason": "No such fetcher'"},
	}
}

// FetchNovel export
func (sf *Silverfish) FetchNovel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	data := map[string]string{}
	decoder.Decode(&data)

	record := (*entity.Novel)(nil)
	if novelID, ok := data["novel_id"]; ok {
		result, err := sf.mgoInf.FindOne(bson.M{"novelID": novelID}, &entity.Novel{})
		if err != nil {
			log.Fatal(fmt.Sprintf("No such NovelID: %s", novelID))
		}
		record = result.(*entity.Novel)
	} else {
		targetURL := data["novel_url"]
		result, err := sf.mgoInf.FindOne(bson.M{"url": targetURL}, &entity.Novel{})
		record = result.(*entity.Novel)

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
