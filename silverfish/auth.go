package silverfish

import (
	"errors"
	"strconv"
	"time"

	entity "silverfish/silverfish/entity"

	"gopkg.in/mgo.v2/bson"
)

// Auth export
type Auth struct {
	hashSalt *string
	userInf  *entity.MongoInf
}

// NewAuth export
func NewAuth(hashSalt *string, userInf *entity.MongoInf) *Auth {
	a := new(Auth)
	a.hashSalt = hashSalt
	a.userInf = userInf
	return a
}

// Register export
func (a *Auth) Register(account, password *string) (*entity.User, error) {
	hashedPassword := SHA512Str(password, a.hashSalt)
	_, err := a.userInf.FindOne(bson.M{"account": *account}, &entity.User{})
	if err != nil {
		if err.Error() == "not found" {
			registerTime := time.Now()
			user := &entity.User{
				Account:           *account,
				Password:          *hashedPassword,
				RegisterDatetime:  registerTime,
				LastLoginDatetime: registerTime,
				Bookmark:          &entity.Bookmark{},
			}
			a.userInf.Upsert(bson.M{"account": *account}, user)
			return &entity.User{
				Account:           user.Account,
				RegisterDatetime:  user.RegisterDatetime,
				LastLoginDatetime: user.LastLoginDatetime,
				Bookmark:          user.Bookmark,
			}, nil
		}
		return nil, err
	}
	return nil, errors.New("account exists")
}

// Login export
func (a *Auth) Login(account, password *string) (*entity.User, error) {
	hashedPassword := SHA512Str(password, a.hashSalt)
	result, err := a.userInf.FindOne(bson.M{"account": *account}, &entity.User{})
	if err != nil {
		if err.Error() == "not found" {
			return nil, errors.New("Account not exists")
		}
		return nil, err
	}
	user := result.(*entity.User)
	if user.Password != *hashedPassword {
		return nil, errors.New("Account or Password wrong")
	}

	user.LastLoginDatetime = time.Now()
	a.userInf.Upsert(bson.M{"account": account}, user)
	return &entity.User{
		Account:           user.Account,
		RegisterDatetime:  user.RegisterDatetime,
		LastLoginDatetime: user.LastLoginDatetime,
		Bookmark:          user.Bookmark,
	}, nil
}

// GetUserBookmark export
func (a *Auth) GetUserBookmark(account *string) (*entity.Bookmark, error) {
	result, err := a.userInf.FindOne(bson.M{"account": account}, &entity.User{})
	if err != nil {
		return nil, errors.New("Account not exists")
	}
	return result.(*entity.User).Bookmark, nil
}

// UpdateBookmark export
func (a *Auth) UpdateBookmark(bookType string, bookID, account, indexStr *string) {
	index, err := strconv.Atoi(*indexStr)
	if err != nil {
		return
	}
	result, err2 := a.userInf.FindOne(bson.M{"account": *account}, &entity.User{})
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
		a.userInf.Upsert(bson.M{"account": *account}, user)
	}
}
