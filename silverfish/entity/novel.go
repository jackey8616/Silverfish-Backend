package entity

import "time"

// NovelInfo export
type NovelInfo struct {
	NovelID       string    `json:"novelID" bson:"novelID"`
	Title         string    `json:"title" bson:"title"`
	Author        string    `json:"author" bson:"author"`
	CoverURL      string    `json:"coverUrl" bson:"coverUrl"`
	LastCrawlTime time.Time `json:"lastCrawlTime" bson:"lastCrawlTime"`
}

// NovelChapter export
type NovelChapter struct {
	Title string `json:"title" bson:"title"`
	URL   string `json:"url" bson:"url"`
}

// Novel export
type Novel struct {
	NovelID       string      	 `json:"novelID" bson:"novelID"`
	DNS           string     	 `json:"dns" bson:"dns"`
	Title         string     	 `json:"title" bson:"title"`
	Author        string     	 `json:"author" bson:"author"`
	Description   string     	 `json:"description" bson:"description"`
	URL           string     	 `json:"url" bson:"url"`
	CoverURL      string 	     `json:"coverUrl" bson:"coverUrl"`
	Chapters      []NovelChapter `json:"chapters" bson:"chapters"`
	LastCrawlTime time.Time 	 `json:"lastCrawlTime" bson:"lastCrawlTime"`
}
