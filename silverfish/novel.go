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

// Novel export
type Novel struct {
	auth          *Auth
	novelInf      *entity.DynamoInf
	novelFetchers map[string]interf.INovelFetcher
	crawlDuration int
}

// NewNovel export
func NewNovel(
	auth *Auth,
	novelInf *entity.DynamoInf,
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

func (n *Novel) findByNovelId(novelId string) (*entity.Novel, error) {
	novelIdString, marshalErr := attributevalue.Marshal(novelId)
	if marshalErr != nil {
		return nil, marshalErr
	}

	keyCond := expression.Key("NovelId").Equal(expression.Value(novelIdString))
	result, err := n.novelInf.FindOne(&keyCond, &entity.Novel{})
	if err != nil {
		return nil, err
	}

	return result.(*entity.Novel), nil
}

func (n *Novel) findByNovelUrl(novelUrl string) (*entity.Novel, error) {
	novelUrlString, marshalErr := attributevalue.Marshal(novelUrl)
	if marshalErr != nil {
		return nil, marshalErr
	}

	keyCond := expression.Key("novelURL").Equal(expression.Value(novelUrlString))
	result, err := n.novelInf.FindOne(&keyCond, &entity.Novel{})
	if err != nil {
		return nil, err
	}

	return result.(*entity.Novel), nil
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
	var keyCond expression.KeyConditionBuilder
	if shouldFetchDisable == false {
		keyCond = expression.Key("IsEnable").Equal(expression.Value(true))
	}

	projConf := expression.NamesList(
		expression.Name("IsEnable"),
		expression.Name("NovelId"),
		expression.Name("CoverUrl"),
		expression.Name("Title"),
		expression.Name("Author"),
		expression.Name("LastCrawlTime"),
	)
	result, err := n.novelInf.FindSelectAll(&keyCond, &projConf, &[]entity.NovelInfo{})
	return result.(*[]entity.NovelInfo), err
}

// GetNovelById export
func (n *Novel) GetNovelById(novelId *string) (*entity.Novel, error) {
	novel, err := n.findByNovelId(*novelId)
	if err != nil {
		return nil, err
	}

	if time.Since(novel.LastCrawlTime).Minutes() > float64(n.crawlDuration) {
		lastCrawlTime := novel.LastCrawlTime
		toUpdateNovel, err := n.novelFetchers[novel.DNS].UpdateNovelInfo(novel)
		if err != nil {
			logrus.Print(err.Error())
			return nil, err
		}

		key, marshalErr := toUpdateNovel.TransformKey()
		if marshalErr != nil {
			return nil, marshalErr
		}
		data := toUpdateNovel.TransformToUpdateBuilder()
		n.novelInf.Update(key, nil, &data)
		logrus.Printf("Updated novel <novel_id: %s, title: %s> since %s", novel.NovelId, novel.Title, lastCrawlTime)
	}

	return novel, nil
}

// RemoveNovelById export
func (n *Novel) RemoveNovelById(novelId *string) error {
	key, marshalErr := attributevalue.MarshalMap(map[string]string{
		"NoveId": *novelId,
	})
	if marshalErr != nil {
		return marshalErr
	}

	err := n.novelInf.Delete(key)
	if err != nil {
		return err
	}

	novelIdBookmarkKey := fmt.Sprintf(`bookmark.novel.%s`, *novelId)
	cond := expression.AttributeExists(expression.Name(novelIdBookmarkKey))
	updateCond := expression.Remove(expression.Name(novelIdBookmarkKey))
	err = n.auth.userInf.Update(nil, &cond, &updateCond)
	if err != nil && err.Error() != "not found" {
		return err
	}
	return nil
}

// AddNovelByURL export
func (n *Novel) AddNovelByURL(novelURL *string) (*entity.Novel, error) {
	novel, err := n.findByNovelUrl(*novelURL)
	if err != nil {
		for _, v := range n.novelFetchers {
			if v.Match(novelURL) {
				record, err := v.CrawlNovel(novelURL)
				if err != nil {
					logrus.Print(err.Error())
					return nil, err
				}
				data, err := attributevalue.MarshalMap(record)
				if err != nil {
					return nil, err
				}
				_, err = n.novelInf.Upsert(data)
				return record, err
			}
		}
		return nil, errors.New("No suit fetcher")

	}

	return novel, nil
}

// GetNovelChapter export
func (n *Novel) GetNovelChapter(novelId, chapterIndex *string) (*string, error) {
	index, err := strconv.Atoi(*chapterIndex)
	if err != nil {
		return nil, errors.New("Invalid chapter index")
	}
	novel, err := n.findByNovelId(*novelId)
	if err != nil {
		return nil, err
	}
	if len((*novel).Chapters) < index {
		return nil, errors.New("Wrong Index")
	} else if val, ok := n.novelFetchers[(*novel).DNS]; ok {
		return val.FetchNovelChapter(novel, index)
	}
	return nil, errors.New("No such fetcher'")
}
