package silverfish

import (
	"errors"
	"strconv"
	"time"

	entity "silverfish/silverfish/entity"

	"gopkg.in/mgo.v2/bson"
)

// User export
type User struct {
	userInf *entity.MongoInf
}

// NewUser export
func NewUser(userInf *entity.MongoInf) *User {
	u := new(User)
	u.userInf = userInf
	return u
}

// GetUser export
func (u *User) GetUser(account *string) (*entity.User, error) {
	result, err := u.userInf.FindOne(bson.M{"account": *account}, &entity.User{})
	if err != nil {
		if err.Error() == "not found" {
			return nil, errors.New("Account not exists")
		}
		return nil, err
	}
	return result.(*entity.User), nil
}

// GetUserBookmark export
func (u *User) GetUserBookmark(account *string) (*entity.Bookmark, error) {
	result, err := u.userInf.FindOne(bson.M{"account": account}, &entity.User{})
	if err != nil {
		return nil, errors.New("Account not exists")
	}
	return result.(*entity.User).Bookmark, nil
}

// UpdateBookmark export
func (u *User) UpdateBookmark(bookType string, bookID, account, indexStr *string) {
	index, err := strconv.Atoi(*indexStr)
	if err != nil {
		return
	}
	result, err2 := u.userInf.FindOne(bson.M{"account": *account}, &entity.User{})
	if err2 == nil {
		user := result.(*entity.User)
		if bookType == "Novel" {
			if val, ok := user.Bookmark.Novel[*bookID]; ok {
				val.LastReadIndex = index
				val.LastReadDatetime = time.Now()
				user.Bookmark.Novel[*bookID] = val
			} else {
				user.Bookmark.Novel[*bookID] = &entity.BookmarkEntry{
					Type:             bookType,
					ID:               *bookID,
					LastReadIndex:    index,
					LastReadDatetime: time.Now(),
				}
			}
		} else {
			if val, ok := user.Bookmark.Comic[*bookID]; ok {
				val.LastReadIndex = index
				val.LastReadDatetime = time.Now()
				user.Bookmark.Comic[*bookID] = val
			} else {
				user.Bookmark.Comic[*bookID] = &entity.BookmarkEntry{
					Type:             bookType,
					ID:               *bookID,
					LastReadIndex:    index,
					LastReadDatetime: time.Now(),
				}
			}
		}
		u.userInf.Upsert(bson.M{"account": *account}, user)
	}
}
