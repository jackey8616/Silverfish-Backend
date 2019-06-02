package entity

import "time"

// APIResponse export
type APIResponse struct {
	Success bool        `json:"success"`
	Fail    bool        `json:"fail"`
	Data    interface{} `json:"data"`
}

// User export
type User struct {
	Account 		  string	`json:"account" bson:"account"`
	Password 		  string	`json:"password" bson:"password"`
	RegisterDatetime  time.Time `json:"registerDatetime" bson:"registerDatetime"`
	LastLoginDatetime time.Time `json:"lastLoginDatetime" bson:"lastLoginDatetime"`
}

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

// ComicInfo export
type ComicInfo struct {
	ComicID       string    `json:"comicID" bson:"comicID"`
	Title         string    `json:"title" bson:"title"`
	Author        string    `json:"author" bson:"author"`
	CoverURL      string    `json:"coverUrl" bson:"coverUrl"`
	LastCrawlTime time.Time `json:"lastCrawlTime" bson:"lastCrawlTime"`
}

// Comic export
type Comic struct {
	ComicID       string         `json:"comicID" bson:"comicID"`
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
	Title 		string		`json:"title" bson:"title"`
	URL 		string		`json:"url" bson:"url"`
	ImageURL	[]string	`json:"imageUrl" bson:"imageUrl"`
}
