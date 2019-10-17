package silverfish

import (
	"errors"
	"log"
	"strconv"
	"time"

	entity "silverfish/silverfish/entity"
	usecase "silverfish/silverfish/usecase"

	"gopkg.in/mgo.v2/bson"
)

// Silverfish export
type Silverfish struct {
	Router        *Router
	Auth          *Auth
	novelInf      *entity.MongoInf
	comicInf      *entity.MongoInf
	novelFetchers map[string]entity.NovelFetcher
	comicFetchers map[string]entity.ComicFetcher
	urls          []string
}

// New export
func New(hashSalt, recaptchaPrivateKey *string, userInf, novelInf, comicInf *entity.MongoInf, urls []string) *Silverfish {
	sf := new(Silverfish)
	sf.Router = NewRouter(recaptchaPrivateKey, sf)
	sf.Auth = NewAuth(hashSalt, userInf)
	sf.novelInf = novelInf
	sf.comicInf = comicInf
	sf.urls = urls
	sf.novelFetchers = map[string]entity.NovelFetcher{
		"www.77xsw.la":      usecase.NewFetcher77xsw("www.77xsw.la"),
		"tw.aixdzs.com":     usecase.NewFetcherAixdzs("tw.aixdzs.com"),
		"tw.hjwzw.com":      usecase.NewFetcherHjwzw("tw.hjwzw.com"),
		"www.biquge.com.cn": usecase.NewFetcherBiquge("www.biquge.com.cn"),
	}
	sf.comicFetchers = map[string]entity.ComicFetcher{
		"www.99comic.co":     usecase.NewFetcher99Comic("www.99comic.co"),
		"www.nokiacn.net":    usecase.NewFetcherNokiacn("www.nokiacn.net"),
		"www.cartoonmad.com": usecase.NewFetcherCartoonmad("www.cartoonmad.com"),
	}
	return sf
}

// getNovels
func (sf *Silverfish) getNovels() (*[]entity.NovelInfo, error) {
	result, err := sf.novelInf.FindSelectAll(nil, bson.M{
		"novelID": 1, "coverUrl": 1, "title": 1, "author": 1, "lastCrawlTime": 1}, &[]entity.NovelInfo{})
	return result.(*[]entity.NovelInfo), err
}

// getNovelByID
func (sf *Silverfish) getNovelByID(novelID *string) (*entity.Novel, error) {
	result, err := sf.novelInf.FindOne(bson.M{"novelID": *novelID}, &entity.Novel{})
	if err != nil {
		return nil, err
	}

	novel := result.(*entity.Novel)
	if time.Since(novel.LastCrawlTime).Hours() > 24 {
		lastCrawlTime := novel.LastCrawlTime
		novel = sf.novelFetchers[novel.DNS].UpdateNovelInfo(novel)
		sf.novelInf.Update(bson.M{"novelID": *novelID}, novel)
		log.Printf("Updated novel <novel_id: %s, title: %s> since %s", novel.NovelID, novel.Title, lastCrawlTime)
	}

	return novel, nil
}

// getNovelByURL
func (sf *Silverfish) getNovelByURL(novelURL *string) (*entity.Novel, error) {
	result, err := sf.novelInf.FindOne(bson.M{"novelURL": *novelURL}, &entity.Novel{})
	if err != nil {
		for _, v := range sf.novelFetchers {
			if v.Match(novelURL) {
				record := v.FetchNovelInfo(novelURL)
				sf.novelInf.Upsert(bson.M{"novelID": record.NovelID}, record)
				return record, nil
			}
		}
		return nil, errors.New("No suit fetcher")

	}

	return result.(*entity.Novel), nil
}

// getNovelChapter
func (sf *Silverfish) getNovelChapter(novelID, chapterIndex *string) (*string, error) {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return nil, errors.New("Invalid chapter index")
	}
	query, err := sf.novelInf.FindOne(bson.M{"novelID": novelID}, &entity.Novel{})
	if err != nil {
		return nil, err
	}
	record := query.(*entity.Novel)
	if len((*record).Chapters) < index {
		return nil, errors.New("Wrong Index")
	} else if val, ok := sf.novelFetchers[(*record).DNS]; ok {
		return val.FetchNovelChapter(record, index), nil
	}
	return nil, errors.New("No such fetcher'")
}

// getComics
func (sf *Silverfish) getComics() (*[]entity.ComicInfo, error) {
	result, err := sf.comicInf.FindSelectAll(nil, bson.M{
		"comicID": 1, "coverUrl": 1, "title": 1, "author": 1, "lastCrawlTime": 1}, &[]entity.ComicInfo{})
	return result.(*[]entity.ComicInfo), err
}

// getComicByID
func (sf *Silverfish) getComicByID(comicID *string) (*entity.Comic, error) {
	result, err := sf.comicInf.FindOne(bson.M{"comicID": *comicID}, &entity.Comic{})
	if err != nil {
		return nil, err
	}

	comic := result.(*entity.Comic)
	if time.Since(comic.LastCrawlTime).Hours() > 24 {
		lastCrawlTime := comic.LastCrawlTime
		comic = sf.comicFetchers[comic.DNS].UpdateComicInfo(comic)
		sf.comicInf.Update(bson.M{"comicID": *comicID}, comic)
		log.Printf("Updated comic <comic_id: %s, title: %s> since %s", comic.ComicID, comic.Title, lastCrawlTime)
	}

	return result.(*entity.Comic), nil
}

// getComicByURL
func (sf *Silverfish) getComicByURL(comicURL *string) (*entity.Comic, error) {
	result, err := sf.comicInf.FindOne(bson.M{"comicURL": *comicURL}, &entity.Comic{})
	if err != nil {
		for _, v := range sf.comicFetchers {
			if v.Match(comicURL) {
				record := v.FetchComicInfo(comicURL)
				sf.comicInf.Upsert(bson.M{"comicID": record.ComicID}, record)
				return record, nil
			}
		}
		return nil, errors.New("No suit fetcher")
	}

	return result.(*entity.Comic), nil
}

// getComicChapter
func (sf *Silverfish) getComicChapter(comicID, chapterIndex *string) ([]string, error) {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return nil, errors.New("Invalid chapter index")
	}
	query, err := sf.comicInf.FindOne(bson.M{"comicID": comicID}, &entity.Comic{})
	if err != nil {
		return nil, err
	}
	record := query.(*entity.Comic)
	if len((*record).Chapters) < index {
		return nil, errors.New("Wrong Index")
	} else if val, ok := sf.comicFetchers[(*record).DNS]; ok {
		if len(record.Chapters[index].ImageURL) == 0 {
			record.Chapters[index].ImageURL = val.FetchComicChapter(record, index)
			sf.comicInf.Update(bson.M{"comicID": record.ComicID}, record)
			log.Printf("Detect <comic:%s> chapter <index: %s/ title: %s> not crawl yet. Crawled.", record.Title, *chapterIndex, record.Chapters[index].Title)
		}
		return record.Chapters[index].ImageURL, nil
	}

	return nil, errors.New("No such fetcher'")
}
