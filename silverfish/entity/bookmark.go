package entity

import "time"

// Bookmark export
type Bookmark struct {
	Novel map[string]*BookmarkEntry `json:"novel" bson:"novel"`
	Comic map[string]*BookmarkEntry `json:"comic" bson:"comic"`
}

// BookmarkEntry export
type BookmarkEntry struct {
	Type             string    `json:"type" bson:"type"`
	ID               string    `json:"ID" bson:"ID"`
	LastReadIndex    int       `json:"lastReadIndex" bson:"lastReadIndex"`
	LastReadDatetime time.Time `json:"lastReadDatetime" bson:"lastReadDatetime"`
}
