package entity

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// NovelInfo export
type NovelInfo struct {
	IsEnable      bool      `json:"IsEnable" bson:"IsEnable"`
	NovelId       string    `json:"NovelId" bson:"NovelId"`
	Title         string    `json:"title" bson:"title"`
	Author        string    `json:"author" bson:"author"`
	Description   string    `json:"description" bson:"description"`
	CoverURL      string    `json:"coverUrl" bson:"coverUrl"`
	LastCrawlTime time.Time `json:"lastCrawlTime" bson:"lastCrawlTime"`
}

// Novel export
type Novel struct {
	Id            string         `json:"id" bson:"id"`
	IsEnable      bool           `json:"IsEnable" bson:"IsEnable"`
	NovelId       string         `json:"NovelId" bson:"NovelId"`
	DNS           string         `json:"dns" bson:"dns"`
	Title         string         `json:"title" bson:"title"`
	Author        string         `json:"author" bson:"author"`
	Description   string         `json:"description" bson:"description"`
	URL           string         `json:"url" bson:"url"`
	CoverURL      string         `json:"coverUrl" bson:"coverUrl"`
	Chapters      []NovelChapter `json:"chapters" bson:"chapters"`
	LastCrawlTime time.Time      `json:"lastCrawlTime" bson:"lastCrawlTime"`
}

// NovelChapter export
type NovelChapter struct {
	Title string `json:"title" bson:"title"`
	URL   string `json:"url" bson:"url"`
}

// GetNovelInfo export
func (novel *Novel) GetNovelInfo() *NovelInfo {
	return &NovelInfo{
		IsEnable:      novel.IsEnable,
		NovelId:       novel.NovelId,
		Title:         novel.Title,
		Author:        novel.Author,
		CoverURL:      novel.CoverURL,
		LastCrawlTime: novel.LastCrawlTime,
	}
}

// SetNovelInfo export
func (novel *Novel) SetNovelInfo(info *NovelInfo) {
	novel.IsEnable = info.IsEnable
	novel.NovelId = info.NovelId
	novel.Title = info.Title
	novel.Author = info.Author
	novel.CoverURL = info.CoverURL
	novel.LastCrawlTime = info.LastCrawlTime
}

func (novel *Novel) TransformKey() (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(map[string]string{
		"NovelId": novel.NovelId,
	})
}

func (novel *Novel) TransformToUpdateBuilder() expression.UpdateBuilder {
	return expression.Set(expression.Name("IsEnable"), expression.Value(novel.IsEnable)).
		Set(expression.Name("DNS"), expression.Value(novel.DNS)).
		Set(expression.Name("Title"), expression.Value(novel.Title)).
		Set(expression.Name("Author"), expression.Value(novel.Author)).
		Set(expression.Name("Description"), expression.Value(novel.Description)).
		Set(expression.Name("URL"), expression.Value(novel.URL)).
		Set(expression.Name("CoverURL"), expression.Value(novel.CoverURL)).
		Set(expression.Name("Chapters"), expression.Value(novel.Chapters)).
		Set(expression.Name("LastCrawlTime"), expression.Value(novel.LastCrawlTime))
}
