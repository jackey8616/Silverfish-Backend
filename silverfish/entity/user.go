package entity

import "time"

// User export
type User struct {
	Id                string
	IsAdmin           bool
	Account           string // Partition Key
	Password          string
	RegisterDatetime  time.Time
	LastLoginDatetime time.Time
	Bookmark          *Bookmark
}
