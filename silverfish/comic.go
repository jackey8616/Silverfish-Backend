package silverfish

import (
	"errors"
	"fmt"
	"silverfish/silverfish/entity"
	interf "silverfish/silverfish/interface"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

// Comic export
type Comic struct {
	auth          *Auth
	comicInf      *entity.MongoInf
	comicFetchers map[string]interf.IComicFetcher
	crawlDuration int
}

// NewComic export
func NewComic(
	auth *Auth,
	comicInf *entity.MongoInf,
	comicFetchers map[string]interf.IComicFetcher,
	crawlDuration int,
) *Comic {
	c := new(Comic)
	c.auth = auth
	c.comicInf = comicInf
	c.comicFetchers = comicFetchers
	c.crawlDuration = crawlDuration
	return c
}

// GetFetchers export
func (c *Comic) GetFetcherNameLists() []string {
	names := []string{}
	for domain := range c.comicFetchers {
		names = append(names, domain)
	}
	return names
}

// GetComics export
func (c *Comic) GetComics(shouldFetchDisable bool) (*[]entity.ComicInfo, error) {
	selector := bson.M{"isEnable": true}
	if shouldFetchDisable == true {
		selector = nil
	}
	result, err := c.comicInf.FindSelectAll(selector, bson.M{
		"comicID": 1, "coverUrl": 1, "title": 1, "author": 1, "lastCrawlTime": 1}, &[]entity.ComicInfo{})
	return result.(*[]entity.ComicInfo), err
}

// GetComicByID export
func (c *Comic) GetComicByID(comicID *string) (*entity.Comic, error) {
	result, err := c.comicInf.FindOne(bson.M{"comicID": *comicID}, &entity.Comic{})
	if err != nil {
		return nil, err
	}

	comic := result.(*entity.Comic)
	if time.Since(comic.LastCrawlTime).Hours() > 24 {
		lastCrawlTime := comic.LastCrawlTime
		comic, err = c.comicFetchers[comic.DNS].UpdateComicInfo(comic)
		if err != nil {
			logrus.Print(err.Error())
			return nil, err
		}
		c.comicInf.Update(bson.M{"comicID": *comicID}, comic)
		logrus.Printf("Updated comic <comic_id: %s, title: %s> since %s", comic.ComicID, comic.Title, lastCrawlTime)
	}

	return result.(*entity.Comic), nil
}

// RemoveComicByID export
func (c *Comic) RemoveComicByID(comicID *string) error {
	err := c.comicInf.Remove(bson.M{"comicID": *comicID})
	if err != nil {
		return err
	}
	err = c.auth.userInf.Update(bson.M{
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
func (c *Comic) AddComicByURL(comicURL *string) (*entity.Comic, error) {
	result, err := c.comicInf.FindOne(bson.M{"comicURL": *comicURL}, &entity.Comic{})
	if err != nil {
		for _, v := range c.comicFetchers {
			if v.Match(comicURL) {
				record, err := v.CrawlComic(comicURL)
				if err != nil {
					logrus.Print(err.Error())
					return nil, err
				}
				c.comicInf.Upsert(bson.M{"comicID": record.ComicID}, record)
				return record, nil
			}
		}
		return nil, errors.New("No suit fetcher")
	}

	return result.(*entity.Comic), nil
}

// GetComicChapter export
func (c *Comic) GetComicChapter(comicID, chapterIndex *string) ([]string, error) {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return nil, errors.New("Invalid chapter index")
	}
	query, err := c.comicInf.FindOne(bson.M{"comicID": comicID}, &entity.Comic{})
	if err != nil {
		return nil, err
	}
	record := query.(*entity.Comic)
	if len((*record).Chapters) < index {
		return nil, errors.New("Wrong Index")
	} else if val, ok := c.comicFetchers[(*record).DNS]; ok {
		if len(record.Chapters[index].ImageURL) == 0 {
			imgURL, err := val.FetchComicChapter(record, index)
			if err != nil {
				logrus.Print(err.Error())
				return nil, err
			}
			record.Chapters[index].ImageURL = imgURL
			c.comicInf.Update(bson.M{"comicID": record.ComicID}, record)
			logrus.Printf("Detect <comic:%s> chapter <index: %s/ title: %s> not crawl yet. Crawled.", record.Title, *chapterIndex, record.Chapters[index].Title)
		}
		return record.Chapters[index].ImageURL, nil
	}

	return nil, errors.New("No such fetcher'")
}
