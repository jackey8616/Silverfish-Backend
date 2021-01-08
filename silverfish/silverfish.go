package silverfish

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	entity "silverfish/silverfish/entity"
	interf "silverfish/silverfish/interface"
	usecase "silverfish/silverfish/usecase"

	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

// Silverfish export
type Silverfish struct {
	Auth          *Auth
	crawlDuration int
	novelInf      *entity.MongoInf
	comicInf      *entity.MongoInf
	novelFetchers map[string]interf.INovelFetcher
	comicFetchers map[string]interf.IComicFetcher
	urls          []string
}

// New export
func New(hashSalt *string, crawlDuration int, userInf, novelInf, comicInf *entity.MongoInf) *Silverfish {
	sf := new(Silverfish)
	sf.Auth = NewAuth(hashSalt, userInf)
	sf.crawlDuration = crawlDuration
	sf.novelInf = novelInf
	sf.comicInf = comicInf
	sf.novelFetchers = map[string]interf.INovelFetcher{
		"www.77xsw.la":      usecase.NewFetcher77xsw("www.77xsw.la"),
		"tw.hjwzw.com":      usecase.NewFetcherHjwzw("tw.hjwzw.com"),
		"www.biquge.com.cn": usecase.NewFetcherBiquge("www.biquge.com.cn"),
		"tw.aixdzs.com":     usecase.NewFetcherAixdzs("tw.aixdzs.com"),
	}
	sf.comicFetchers = map[string]interf.IComicFetcher{
		//"www.99comic.co":     usecase.NewFetcher99Comic("www.99comic.co"),
		"www.nokiacn.net":    usecase.NewFetcherNokiacn("www.nokiacn.net"),
		"www.cartoonmad.com": usecase.NewFetcherCartoonmad("www.cartoonmad.com"),
		"comicbus.com":       usecase.NewFetcherComicbus("comicbus.com"),
		"www.manhuaniu.com":  usecase.NewFetcherManhuaniu("www.manhuaniu.com"),
		"www.mangabz.com":    usecase.NewFetcherMangabz("www.mangabz.com"),
		"m.happymh.com":      usecase.NewFetcherHappymh("m.happymh.com"),
	}
	return sf
}

// GetLists export
func (sf *Silverfish) GetLists() map[string][]string {
	lists := map[string][]string{
		"novels": []string{},
		"comics": []string{},
	}
	for domain := range sf.novelFetchers {
		lists["novels"] = append(lists["novels"], domain)
	}
	for domain := range sf.comicFetchers {
		lists["comics"] = append(lists["comics"], domain)
	}
	return lists
}

// GetNovels export
func (sf *Silverfish) GetNovels() (*[]entity.NovelInfo, error) {
	result, err := sf.novelInf.FindSelectAll(nil, bson.M{
		"novelID": 1, "coverUrl": 1, "title": 1, "author": 1, "lastCrawlTime": 1}, &[]entity.NovelInfo{})
	return result.(*[]entity.NovelInfo), err
}

// GetNovelByID export
func (sf *Silverfish) GetNovelByID(novelID *string) (*entity.Novel, error) {
	result, err := sf.novelInf.FindOne(bson.M{"novelID": *novelID}, &entity.Novel{})
	if err != nil {
		return nil, err
	}

	novel := result.(*entity.Novel)
	if time.Since(novel.LastCrawlTime).Minutes() > float64(sf.crawlDuration) {
		lastCrawlTime := novel.LastCrawlTime
		novel, err = sf.novelFetchers[novel.DNS].UpdateNovelInfo(novel)
		if err != nil {
			logrus.Print(err.Error())
			return nil, err
		}
		sf.novelInf.Update(bson.M{"novelID": *novelID}, novel)
		logrus.Printf("Updated novel <novel_id: %s, title: %s> since %s", novel.NovelID, novel.Title, lastCrawlTime)
	}

	return novel, nil
}

// RemoveNovelByID export
func (sf *Silverfish) RemoveNovelByID(novelID *string) error {
	err := sf.novelInf.Remove(bson.M{"novelID": *novelID})
	if err != nil {
		return err
	}
	err = sf.Auth.userInf.Update(bson.M{
		fmt.Sprintf(`bookmark.novel.%s`, *novelID): bson.M{
			"$exists": true,
		},
	}, bson.M{
		"$unset": bson.M{
			fmt.Sprintf(`bookmark.novel.%s`, *novelID): "",
		},
	})
	if err != nil && err.Error() != "not found" {
		return err
	}
	return nil
}

// AddNovelByURL export
func (sf *Silverfish) AddNovelByURL(novelURL *string) (*entity.Novel, error) {
	result, err := sf.novelInf.FindOne(bson.M{"novelURL": *novelURL}, &entity.Novel{})
	if err != nil {
		for _, v := range sf.novelFetchers {
			if v.Match(novelURL) {
				record, err := v.CrawlNovel(novelURL)
				if err != nil {
					logrus.Print(err.Error())
					return nil, err
				}
				sf.novelInf.Upsert(bson.M{"novelID": record.NovelID}, record)
				return record, nil
			}
		}
		return nil, errors.New("No suit fetcher")

	}

	return result.(*entity.Novel), nil
}

// GetNovelChapter export
func (sf *Silverfish) GetNovelChapter(novelID, chapterIndex *string) (*string, error) {
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
		return val.FetchNovelChapter(record, index)
	}
	return nil, errors.New("No such fetcher'")
}

// GetComics export
func (sf *Silverfish) GetComics() (*[]entity.ComicInfo, error) {
	result, err := sf.comicInf.FindSelectAll(nil, bson.M{
		"comicID": 1, "coverUrl": 1, "title": 1, "author": 1, "lastCrawlTime": 1}, &[]entity.ComicInfo{})
	return result.(*[]entity.ComicInfo), err
}

// GetComicByID export
func (sf *Silverfish) GetComicByID(comicID *string) (*entity.Comic, error) {
	result, err := sf.comicInf.FindOne(bson.M{"comicID": *comicID}, &entity.Comic{})
	if err != nil {
		return nil, err
	}

	comic := result.(*entity.Comic)
	if time.Since(comic.LastCrawlTime).Hours() > 24 {
		lastCrawlTime := comic.LastCrawlTime
		comic, err = sf.comicFetchers[comic.DNS].UpdateComicInfo(comic)
		if err != nil {
			logrus.Print(err.Error())
			return nil, err
		}
		sf.comicInf.Update(bson.M{"comicID": *comicID}, comic)
		logrus.Printf("Updated comic <comic_id: %s, title: %s> since %s", comic.ComicID, comic.Title, lastCrawlTime)
	}

	return result.(*entity.Comic), nil
}

// RemoveComicByID export
func (sf *Silverfish) RemoveComicByID(comicID *string) error {
	err := sf.comicInf.Remove(bson.M{"comicID": *comicID})
	if err != nil {
		return err
	}
	err = sf.Auth.userInf.Update(bson.M{
		fmt.Sprintf(`bookmark.comic.%s`, *comicID): bson.M{
			"$exists": true,
		},
	}, bson.M{
		"$unset": bson.M{
			fmt.Sprintf(`bookmark.comic.%s`, *comicID): "",
		},
	})
	if err != nil && err.Error() != "not found" {
		return err
	}
	return nil
}

// AddComicByURL export
func (sf *Silverfish) AddComicByURL(comicURL *string) (*entity.Comic, error) {
	result, err := sf.comicInf.FindOne(bson.M{"comicURL": *comicURL}, &entity.Comic{})
	if err != nil {
		for _, v := range sf.comicFetchers {
			if v.Match(comicURL) {
				record, err := v.CrawlComic(comicURL)
				if err != nil {
					logrus.Print(err.Error())
					return nil, err
				}
				sf.comicInf.Upsert(bson.M{"comicID": record.ComicID}, record)
				return record, nil
			}
		}
		return nil, errors.New("No suit fetcher")
	}

	return result.(*entity.Comic), nil
}

// GetComicChapter export
func (sf *Silverfish) GetComicChapter(comicID, chapterIndex *string) ([]string, error) {
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
			imgURL, err := val.FetchComicChapter(record, index)
			if err != nil {
				logrus.Print(err.Error())
				return nil, err
			}
			record.Chapters[index].ImageURL = imgURL
			sf.comicInf.Update(bson.M{"comicID": record.ComicID}, record)
			logrus.Printf("Detect <comic:%s> chapter <index: %s/ title: %s> not crawl yet. Crawled.", record.Title, *chapterIndex, record.Chapters[index].Title)
		}
		return record.Chapters[index].ImageURL, nil
	}

	return nil, errors.New("No such fetcher'")
}
