package entity

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// ComicInfo export
type ComicInfo struct {
	IsEnable      bool      `json:"isEnable" bson:"isEnable"`
	ComicId       string    `json:"comicId" bson:"comicId"`
	Title         string    `json:"title" bson:"title"`
	Author        string    `json:"author" bson:"author"`
	Description   string    `json:"description" bson:"description"`
	CoverURL      string    `json:"coverUrl" bson:"coverUrl"`
	LastCrawlTime time.Time `json:"lastCrawlTime" bson:"lastCrawlTime"`
}

// Comic export
type Comic struct {
	IsEnable      bool           `json:"isEnable" bson:"isEnable"`
	ComicId       string         `json:"comicId" bson:"comicId"`
	DNS           string         `json:"dns" bson:"dns"`
	Title         string         `json:"title" bson:"title"`
	Author        string         `json:"author" bson:"author"`
	Description   string         `json:"description" bson:"description"`
	URL           string         `json:"url" bson:"url"`
	CoverURL      string         `json:"coverUrl" bson:"coverUrl"`
	Chapters      []ComicChapter `json:"chapters" bson:"chapters"`
	LastCrawlTime time.Time      `json:"lastCrawlTime" bson:"lastCrawlTime"`
}

// ComicChapter export
type ComicChapter struct {
	Title    string   `json:"title" bson:"title"`
	URL      string   `json:"url" bson:"url"`
	ImageURL []string `json:"imageUrl" bson:"imageUrl"`
}

// GetComicInfo export
func (comic *Comic) GetComicInfo() *ComicInfo {
	return &ComicInfo{
		IsEnable:      comic.IsEnable,
		ComicId:       comic.ComicId,
		Title:         comic.Title,
		Author:        comic.Author,
		Description:   comic.Description,
		CoverURL:      comic.CoverURL,
		LastCrawlTime: comic.LastCrawlTime,
	}
}

// SetComicInfo export
func (comic *Comic) SetComicInfo(info *ComicInfo) {
	comic.IsEnable = info.IsEnable
	comic.ComicId = info.ComicId
	comic.Title = info.Title
	comic.Author = info.Author
	comic.Description = info.Description
	comic.CoverURL = info.CoverURL
	comic.LastCrawlTime = info.LastCrawlTime
}

func (comic *Comic) TransformKey() (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(map[string]string{
		"comicId": comic.ComicId,
	})
}

func (comic *Comic) TransformToUpdateBuilder() expression.UpdateBuilder {
	return expression.Set(expression.Name("IsEnable"), expression.Value(comic.IsEnable)).
		Set(expression.Name("DNS"), expression.Value(comic.DNS)).
		Set(expression.Name("Title"), expression.Value(comic.Title)).
		Set(expression.Name("Author"), expression.Value(comic.Author)).
		Set(expression.Name("Description"), expression.Value(comic.Description)).
		Set(expression.Name("URL"), expression.Value(comic.URL)).
		Set(expression.Name("CoverURL"), expression.Value(comic.CoverURL)).
		Set(expression.Name("Chapters"), expression.Value(comic.Chapters)).
		Set(expression.Name("LastCrawlTime"), expression.Value(comic.LastCrawlTime))
}
