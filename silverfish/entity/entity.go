package entity

import "time"

// User export
type User struct {
	IsAdmin           bool      `json:"isAdmin" bson:"isAdmin"`
	Account           string    `json:"account" bson:"account"`
	Password          string    `json:"password" bson:"password"`
	RegisterDatetime  time.Time `json:"registerDatetime" bson:"registerDatetime"`
	LastLoginDatetime time.Time `json:"lastLoginDatetime" bson:"lastLoginDatetime"`
	Bookmark          *Bookmark `json:"bookmark" bson:"bookmark"`
}

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
