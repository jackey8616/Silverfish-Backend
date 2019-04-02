package entity

import "time"

// Chapter export
type Chapter struct {
	Title string `json:"title" bson:"title"`
	URL   string `json:"url" bson:"url"`
}

// Novel export
type Novel struct {
	NovelID       string    `json:"novelID" bson:"novelID"`
	DNS           string    `json:"dns" bson:"dns"`
	Title         string    `json:"title" bson:"title"`
	Author        string    `json:"author" bson:"author"`
	Description   string    `json:"description" bson:"description"`
	URL           string    `json:"url" bson:"url"`
	CoverURL      string    `json:"coverUrl" bson:"coverUrl"`
	Chapters      []Chapter `json:"chapters" bson:"chapters"`
	LastCrawlTime time.Time `json:"lastCrawlTime" bson:"lastCrawlTime"`
}
