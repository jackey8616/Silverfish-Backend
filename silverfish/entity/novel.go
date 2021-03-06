package entity

import "time"

// NovelInfo export
type NovelInfo struct {
	IsEnable      bool      `json:"isEnable" bson:"isEnable"`
	NovelID       string    `json:"novelID" bson:"novelID"`
	Title         string    `json:"title" bson:"title"`
	Author        string    `json:"author" bson:"author"`
	Description   string    `json:"description" bson:"description"`
	CoverURL      string    `json:"coverUrl" bson:"coverUrl"`
	LastCrawlTime time.Time `json:"lastCrawlTime" bson:"lastCrawlTime"`
}

// Novel export
type Novel struct {
	IsEnable      bool           `json:"isEnable" bson:"isEnable"`
	NovelID       string         `json:"novelID" bson:"novelID"`
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
		NovelID:       novel.NovelID,
		Title:         novel.Title,
		Author:        novel.Author,
		CoverURL:      novel.CoverURL,
		LastCrawlTime: novel.LastCrawlTime,
	}
}

// SetNovelInfo export
func (novel *Novel) SetNovelInfo(info *NovelInfo) {
	novel.IsEnable = info.IsEnable
	novel.NovelID = info.NovelID
	novel.Title = info.Title
	novel.Author = info.Author
	novel.CoverURL = info.CoverURL
	novel.LastCrawlTime = info.LastCrawlTime
}
