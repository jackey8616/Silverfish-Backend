package entity

import "time"

// APIResponse export
type APIResponse struct {
	Success bool        `json:"success"`
	Fail    bool        `json:"fail"`
	Data    interface{} `json:"data"`
}

// NovelInfo export
type NovelInfo struct {
	NovelID       string    `json:"novelID" bson:"novelID"`
	Title         string    `json:"title" bson:"title"`
	CoverURL      string    `json:"coverUrl" bson:"coverUrl"`
	LastCrawlTime time.Time `json:"lastCrawlTime" bson:"lastCrawlTime"`
}

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
