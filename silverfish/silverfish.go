package silverfish

import (
	"log"
	"strconv"

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
		"tw.hjwzw.com": usecase.NewFetcherHjwzw("tw.hjwzw.com"),
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

// getNovelByID
func (sf *Silverfish) getNovelByID(novelID *string) *entity.APIResponse {
	result, err := sf.mgoInf.FindOne(bson.M{"novelID": *novelID}, &entity.Novel{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	novel = result.(*entity.Novel)
	if time.Now().Sub(novel.LatCrawlTime).Days() > 1 {
		log.Printf("Updating novel <novel_id: %s, title: %s>", novel.NovelID, novel.Title)
		novel = sf.fetchers[result.DNS].UpdateNovelInfo(novel)
		sf.mgoInf.Update(bson.M{"novelID": *novelID}, novel)
	}
	return &entity.APIResponse{
		Success: true,
		Data:    novel,
	}
}

// getNovelByURL
func (sf *Silverfish) getNovelByURL(novelURL *string) *entity.APIResponse {
	result, err := sf.mgoInf.FindOne(bson.M{"novelURL": *novelURL}, &entity.Novel{})
	if err != nil {
		for _, v := range sf.fetchers {
			if v.Match(novelURL) {
				record := v.FetchNovelInfo(novelURL)
				sf.mgoInf.Upsert(bson.M{"novelID": record.NovelID}, record)
				return &entity.APIResponse{
					Success: true,
					Data:    record,
				}
			}
		}
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": "No suit fetcher"},
		}
	}

	return &entity.APIResponse{
		Success: true,
		Data:    result.(*entity.Novel),
	}
}

// getChapter
func (sf *Silverfish) getChapter(novelID, chapterIndex *string) *entity.APIResponse {
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
	} else if val, ok := sf.fetchers[(*record).DNS]; ok {
		return &entity.APIResponse{
			Success: true,
			Data:    val.FetchChapter(record, index),
		}
	}
	return &entity.APIResponse{
		Success: true,
		Data:    map[string]string{"reason": "No such fetcher'"},
	}
}
