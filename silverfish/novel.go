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

// Novel export
type Novel struct {
	auth          *Auth
	novelInf      *entity.MongoInf
	novelFetchers map[string]interf.INovelFetcher
	crawlDuration int
}

// NewNovel export
func NewNovel(
	auth *Auth,
	novelInf *entity.MongoInf,
	novelFetchers map[string]interf.INovelFetcher,
	crawlDuration int,
) *Novel {
	n := new(Novel)
	n.auth = auth
	n.novelInf = novelInf
	n.novelFetchers = novelFetchers
	n.crawlDuration = crawlDuration
	return n
}

// GetFetchers export
func (n *Novel) GetFetcherNameLists() []string {
	names := []string{}
	for domain := range n.novelFetchers {
		names = append(names, domain)
	}
	return names
}

// GetNovels export
func (n *Novel) GetNovels(shouldFetchDisable bool) (*[]entity.NovelInfo, error) {
	selector := bson.M{"isEnable": true}
	if shouldFetchDisable {
		selector = nil
	}
	result, err := n.novelInf.FindSelectAll(selector, bson.M{
		"novelID": 1, "coverUrl": 1, "title": 1, "author": 1, "lastCrawlTime": 1}, &[]entity.NovelInfo{})
	return result.(*[]entity.NovelInfo), err
}

// GetNovelByID export
func (n *Novel) GetNovelByID(novelID *string) (*entity.Novel, error) {
	result, err := n.novelInf.FindOne(bson.M{"novelID": *novelID}, &entity.Novel{})
	if err != nil {
		return nil, err
	}

	novel := result.(*entity.Novel)
	if time.Since(novel.LastCrawlTime).Minutes() > float64(n.crawlDuration) {
		lastCrawlTime := novel.LastCrawlTime
		novel, err = n.novelFetchers[novel.DNS].UpdateNovelInfo(novel)
		if err != nil {
			logrus.Print(err.Error())
			return nil, err
		}
		n.novelInf.Update(bson.M{"novelID": *novelID}, novel)
		logrus.Printf("Updated novel <novel_id: %s, title: %s> since %s", novel.NovelID, novel.Title, lastCrawlTime)
	}

	return novel, nil
}

// RemoveNovelByID export
func (n *Novel) RemoveNovelByID(novelID *string) error {
	err := n.novelInf.Remove(bson.M{"novelID": *novelID})
	if err != nil {
		return err
	}
	err = n.auth.userInf.Update(bson.M{
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
func (n *Novel) AddNovelByURL(novelURL *string) (*entity.Novel, error) {
	result, err := n.novelInf.FindOne(bson.M{"novelURL": *novelURL}, &entity.Novel{})
	if err != nil {
		for _, v := range n.novelFetchers {
			if v.Match(novelURL) {
				record, err := v.CrawlNovel(novelURL)
				if err != nil {
					logrus.Print(err.Error())
					return nil, err
				}
				n.novelInf.Upsert(bson.M{"novelID": record.NovelID}, record)
				return record, nil
			}
		}
		return nil, errors.New("No suit fetcher")

	}

	return result.(*entity.Novel), nil
}

// GetNovelChapter export
func (n *Novel) GetNovelChapter(novelID, chapterIndex *string) (*string, error) {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return nil, errors.New("Invalid chapter index")
	}
	query, err := n.novelInf.FindOne(bson.M{"novelID": novelID}, &entity.Novel{})
	if err != nil {
		return nil, err
	}
	record := query.(*entity.Novel)
	if len((*record).Chapters) < index {
		return nil, errors.New("Wrong Index")
	} else if val, ok := n.novelFetchers[(*record).DNS]; ok {
		return val.FetchNovelChapter(record, index)
	}
	return nil, errors.New("No such fetcher'")
}
