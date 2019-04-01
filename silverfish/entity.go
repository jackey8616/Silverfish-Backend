package silverfish

import "time"

// Chapter export
type Chapter struct {
	Title string `json:"title" bson:"title"`
	URL   string `json:"url" bson:"url"`
}

// Novel export
type Novel struct {
	Title         string    `json:"title" bson:"title"`
	DNS           string    `json:"dns" bson:"dns"`
	URL           string    `json:"url" bson:"url"`
	CoverURL      string    `json:"coverUrl" bson:"coverUrl"`
	Chapters      []Chapter `json:"chapters" bson:"chapters"`
	LastCrawlTime time.Time `json:"lastCrawlTime" bson:"lastCrawlTime"`
}
