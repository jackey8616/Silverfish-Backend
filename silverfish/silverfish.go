package silverfish

import (
	"time"
	"log"
	"strconv"

	"github.com/jackey8616/Silverfish-backend/silverfish/entity"
	"github.com/jackey8616/Silverfish-backend/silverfish/usecase"
	"gopkg.in/mgo.v2/bson"
)

// Silverfish export
type Silverfish struct {
	Router        *Router
	novelInf      *entity.MongoInf
	comicInf      *entity.MongoInf
	novelFetchers map[string]entity.NovelFetcher
	comicFetchers map[string]entity.ComicFetcher
	urls          []string
}

// New export
func New(novelInf, comicInf *entity.MongoInf, urls []string) *Silverfish {
	sf := new(Silverfish)
	sf.Router = NewRouter(sf)
	sf.novelInf = novelInf
	sf.comicInf = comicInf
	sf.urls = urls
	sf.novelFetchers = map[string]entity.NovelFetcher{
		"www.77xsw.la": usecase.NewFetcher77xsw("www.77xsw.la"),
		"tw.hjwzw.com": usecase.NewFetcherHjwzw("tw.hjwzw.com"),
	}
	sf.comicFetchers = map[string]entity.ComicFetcher{
		"www.99comic.co": usecase.NewFetcher99Comic("www.99comic.co"),
		"www.nokiacn.net": usecase.NewFetcherNokiacn("www.nokiacn.net"),
	}
	return sf
}

// getNovels
func (sf *Silverfish) getNovels() *entity.APIResponse {
	result, err := sf.novelInf.FindSelectAll(nil, bson.M{
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
	result, err := sf.novelInf.FindOne(bson.M{"novelID": *novelID}, &entity.Novel{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	
	novel := result.(*entity.Novel)
	if time.Since(novel.LastCrawlTime).Hours() > 24 {
		lastCrawlTime := novel.LastCrawlTime
		novel = sf.novelFetchers[novel.DNS].UpdateNovelInfo(novel)
		sf.novelInf.Update(bson.M{"novelID": *novelID}, novel)
		log.Printf("Updated novel <novel_id: %s, title: %s> since %s", novel.NovelID, novel.Title, lastCrawlTime)
	}

	return &entity.APIResponse{
		Success: true,
		Data:    novel,
	}
}

// getNovelByURL
func (sf *Silverfish) getNovelByURL(novelURL *string) *entity.APIResponse {
	result, err := sf.novelInf.FindOne(bson.M{"novelURL": *novelURL}, &entity.Novel{})
	if err != nil {
		for _, v := range sf.novelFetchers {
			if v.Match(novelURL) {
				record := v.FetchNovelInfo(novelURL)
				sf.novelInf.Upsert(bson.M{"novelID": record.NovelID}, record)
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

// getNovelChapter
func (sf *Silverfish) getNovelChapter(novelID, chapterIndex *string) *entity.APIResponse {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": "Invalid chapter index"},
		}
	}
	query, err := sf.novelInf.FindOne(bson.M{"novelID": novelID}, &entity.Novel{})
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
	} else if val, ok := sf.novelFetchers[(*record).DNS]; ok {
		return &entity.APIResponse{
			Success: true,
			Data:    val.FetchNovelChapter(record, index),
		}
	}
	return &entity.APIResponse{
		Success: true,
		Data:    map[string]string{"reason": "No such fetcher'"},
	}
}

// getComics
func (sf *Silverfish) getComics() *entity.APIResponse {
	result, err := sf.comicInf.FindSelectAll(nil, bson.M{
		"comicID": 1, "coverUrl": 1, "title": 1, "lastCrawlTime": 1}, &[]entity.ComicInfo{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	return &entity.APIResponse{
		Success: true,
		Data: 	 result.(*[]entity.ComicInfo),
	}
}

// getComicByID
func (sf *Silverfish) getComicByID(comicID *string) *entity.APIResponse {
	result, err := sf.comicInf.FindOne(bson.M{"comicID": *comicID}, &entity.Comic{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	return &entity.APIResponse{
		Success: true,
		Data:	 result.(*entity.Comic),
	}
}

// getComicByURL
func (sf *Silverfish) getComicByURL(comicURL *string) *entity.APIResponse {
	result, err := sf.comicInf.FindOne(bson.M{"comicURL": *comicURL}, &entity.Comic{})
	if err != nil {
		for _, v := range sf.comicFetchers {
			if v.Match(comicURL) {
				record := v.FetchComicInfo(comicURL)
				sf.comicInf.Upsert(bson.M{"comicID": record.ComicID}, record)
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
		Data:    result.(*entity.Comic),
	}
}

// getComicChapter
func (sf *Silverfish) getComicChapter(comicID, chapterIndex *string) *entity.APIResponse {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": "Invalid chapter index"},
		}
	}
	query, err := sf.comicInf.FindOne(bson.M{"comicID": comicID}, &entity.Comic{})
	if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	record := query.(*entity.Comic)
	if len((*record).Chapters) < index {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": "Wrong Index"},
		}
	} else if val, ok := sf.comicFetchers[(*record).DNS]; ok {
		if len(record.Chapters[index].ImageURL) == 0 {
			record.Chapters[index].ImageURL = val.FetchComicChapter(record, index)
			sf.comicInf.Update(bson.M{"comicID": record.ComicID}, record)
			log.Printf("Detect <comic:%s> chapter <index: %s/ title: %s> not crawl yet. Crawled.", record.Title, *chapterIndex, record.Chapters[index].Title)
		}
		return &entity.APIResponse{
			Success: true,
			Data:    record.Chapters[index].ImageURL,
		}
	}
	return &entity.APIResponse{
		Success: true,
		Data:    map[string]string{"reason": "No such fetcher'"},
	}
}
