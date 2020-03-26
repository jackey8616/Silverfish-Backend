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
