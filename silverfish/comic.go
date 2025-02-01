package silverfish

import (
	"errors"
	"fmt"
	"silverfish/silverfish/entity"
	interf "silverfish/silverfish/interface"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/sirupsen/logrus"
)

// Comic export
type Comic struct {
	auth          *Auth
	comicInf      *entity.DynamoInf
	comicFetchers map[string]interf.IComicFetcher
	crawlDuration int
}

// NewComic export
func NewComic(
	auth *Auth,
	comicInf *entity.DynamoInf,
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

func (c *Comic) findByComicId(comicId string) (*entity.Comic, error) {
	comicIdString, marshalErr := attributevalue.Marshal(comicId)
	if marshalErr != nil {
		return nil, marshalErr
	}

	keyCond := expression.Key("ComicId").Equal(expression.Value(comicIdString))
	result, err := c.comicInf.FindOne(&keyCond, &entity.Comic{})
	if err != nil {
		return nil, err
	}

	return result.(*entity.Comic), nil
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
	var keyCond expression.KeyConditionBuilder
	if shouldFetchDisable == false {
		keyCond = expression.Key("IsEnable").Equal(expression.Value(true))
	}

	projConf := expression.NamesList(
		expression.Name("IsEnable"),
		expression.Name("ComicId"),
		expression.Name("CoverUrl"),
		expression.Name("Title"),
		expression.Name("Author"),
		expression.Name("LastCrawlTime"),
	)

	result, err := c.comicInf.FindSelectAll(&keyCond, &projConf, &[]entity.ComicInfo{})
	return result.(*[]entity.ComicInfo), err
}

func (c *Comic) findByComicUrl(comicUrl string) (*entity.Comic, error) {
	comicUrlString, marshalErr := attributevalue.Marshal(comicUrl)
	if marshalErr != nil {
		return nil, marshalErr
	}

	keyCond := expression.Key("comicURL").Equal(expression.Value(comicUrlString))
	result, err := c.comicInf.FindOne(&keyCond, &entity.Comic{})
	if err != nil {
		return nil, err
	}

	return result.(*entity.Comic), nil
}

// GetComicById export
func (c *Comic) GetComicById(comicId *string) (*entity.Comic, error) {
	comic, err := c.findByComicId(*comicId)
	if err != nil {
		return nil, err
	}

	if time.Since(comic.LastCrawlTime).Hours() > 24 {
		lastCrawlTime := comic.LastCrawlTime
		toUpdateComic, err := c.comicFetchers[comic.DNS].UpdateComicInfo(comic)
		if err != nil {
			logrus.Print(err.Error())
			return nil, err
		}
		key, marshalErr := toUpdateComic.TransformKey()
		if marshalErr != nil {
			return nil, marshalErr
		}
		data := toUpdateComic.TransformToUpdateBuilder()
		c.comicInf.Update(key, nil, &data)
		logrus.Printf("Updated comic <comic_id: %s, title: %s> since %s", comic.ComicId, comic.Title, lastCrawlTime)
	}

	return comic, nil
}

// RemoveComicById export
func (c *Comic) RemoveComicById(comicId *string) error {
	key, marshalErr := attributevalue.MarshalMap(map[string]string{
		"ComicId": *comicId,
	})
	if marshalErr != nil {
		return marshalErr
	}

	err := c.comicInf.Delete(key)
	if err != nil {
		return err
	}

	comicIdBookmarkKey := fmt.Sprintf(`bookmark.comic.%s`, *comicId)
	cond := expression.AttributeExists(expression.Name(comicIdBookmarkKey))
	updateCond := expression.Remove(expression.Name(comicIdBookmarkKey))
	err = c.auth.userInf.Update(nil, &cond, &updateCond)
	if err != nil && err.Error() != "not found" {
		return err
	}
	return nil
}

// AddComicByURL export
func (c *Comic) AddComicByURL(comicURL *string) (*entity.Comic, error) {
	comic, err := c.findByComicUrl(*comicURL)
	if err != nil {
		for _, v := range c.comicFetchers {
			if v.Match(comicURL) {
				record, err := v.CrawlComic(comicURL)
				if err != nil {
					logrus.Print(err.Error())
					return nil, err
				}
				data, err := attributevalue.MarshalMap(record)
				if err != nil {
					return nil, err
				}
				c.comicInf.Upsert(data)
				return record, nil
			}
		}
		return nil, errors.New("No suit fetcher")
	}

	return comic, nil
}

// GetComicChapter export
func (c *Comic) GetComicChapter(comicId, chapterIndex *string) ([]string, error) {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return nil, errors.New("Invalid chapter index")
	}
	comic, err := c.findByComicId(*comicId)
	if err != nil {
		return nil, err
	}
	if len((*comic).Chapters) < index {
		return nil, errors.New("Wrong Index")
	} else if val, ok := c.comicFetchers[(*comic).DNS]; ok {
		if len(comic.Chapters[index].ImageURL) == 0 {
			imgURL, err := val.FetchComicChapter(comic, index)
			if err != nil {
				logrus.Print(err.Error())
				return nil, err
			}
			comic.Chapters[index].ImageURL = imgURL
			key, marshalErr := comic.TransformKey()
			if marshalErr != nil {
				return nil, marshalErr
			}
			data := comic.TransformToUpdateBuilder()
			c.comicInf.Update(key, nil, &data)
			logrus.Printf("Detect <comic:%s> chapter <index: %s/ title: %s> not crawl yet. Crawled.", comic.Title, *chapterIndex, comic.Chapters[index].Title)
		}
		return comic.Chapters[index].ImageURL, nil
	}

	return nil, errors.New("No such fetcher'")
}
