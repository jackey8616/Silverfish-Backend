package entity

import "time"

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
