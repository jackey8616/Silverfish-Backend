package entity

import "time"

// ComicInfo export
type ComicInfo struct {
	ComicID       string    `json:"comicID" bson:"comicID"`
	Title         string    `json:"title" bson:"title"`
	Author        string    `json:"author" bson:"author"`
	Description   string    `json:"description" bson:"description"`
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
	Title    string   `json:"title" bson:"title"`
	URL      string   `json:"url" bson:"url"`
	ImageURL []string `json:"imageUrl" bson:"imageUrl"`
}

// GetComicInfo export
func (comic *Comic) GetComicInfo() *ComicInfo {
	return &ComicInfo{
		ComicID:       comic.ComicID,
		Title:         comic.Title,
		Author:        comic.Author,
		Description:   comic.Description,
		CoverURL:      comic.CoverURL,
		LastCrawlTime: comic.LastCrawlTime,
	}
}

// SetComicInfo export
func (comic *Comic) SetComicInfo(info *ComicInfo) {
	comic.ComicID = info.ComicID
	comic.Title = info.Title
	comic.Author = info.Author
	comic.Description = info.Description
	comic.CoverURL = info.CoverURL
	comic.LastCrawlTime = info.LastCrawlTime
}
