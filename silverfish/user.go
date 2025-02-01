package silverfish

import (
	"errors"
	"strconv"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
)

// User export
type User struct {
	userInf *entity.DynamoInf
}

// NewUser export
func NewUser(userInf *entity.DynamoInf) *User {
	u := new(User)
	u.userInf = userInf
	return u
}

func (u *User) findByAccount(account string) (*entity.User, error) {
	accountString, marshalErr := attributevalue.Marshal(account)
	if marshalErr != nil {
		return nil, marshalErr
	}

	keyCond := expression.Key("Account").Equal(expression.Value(accountString))
	result, err := u.userInf.FindOne(&keyCond, &entity.User{})
	if err != nil {
		return nil, err
	}

	return result.(*entity.User), nil
}

func (u *User) upsertUser(user *entity.User) (*entity.User, error) {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, err
	}

	_, putErr := u.userInf.Upsert(item)
	return user, putErr
}

// GetUser export
func (u *User) GetUser(account *string) (*entity.User, error) {
	user, err := u.findByAccount(*account)
	if err != nil {
		if err.Error() == "not found" {
			return nil, errors.New("Account not exists")
		}
		return nil, err
	}
	return user, nil
}

// GetUserBookmark export
func (u *User) GetUserBookmark(account *string) (*entity.Bookmark, error) {
	user, err := u.findByAccount(*account)
	if err != nil {
		return nil, errors.New("Account not exists")
	}
	return user.Bookmark, nil
}

// UpdateBookmark export
func (u *User) UpdateBookmark(bookType string, bookId, account, indexStr *string) {
	index, err := strconv.Atoi(*indexStr)
	if err != nil {
		return
	}
	user, err2 := u.findByAccount(*account)
	if err2 == nil {
		if bookType == "Novel" {
			if val, ok := user.Bookmark.Novel[*bookId]; ok {
				val.LastReadIndex = index
				val.LastReadDatetime = time.Now()
				user.Bookmark.Novel[*bookId] = val
			} else {
				user.Bookmark.Novel[*bookId] = &entity.BookmarkEntry{
					Type:             bookType,
					Id:               *bookId,
					LastReadIndex:    index,
					LastReadDatetime: time.Now(),
				}
			}
		} else {
			if val, ok := user.Bookmark.Comic[*bookId]; ok {
				val.LastReadIndex = index
				val.LastReadDatetime = time.Now()
				user.Bookmark.Comic[*bookId] = val
			} else {
				user.Bookmark.Comic[*bookId] = &entity.BookmarkEntry{
					Type:             bookType,
					Id:               *bookId,
					LastReadIndex:    index,
					LastReadDatetime: time.Now(),
				}
			}
		}
		u.upsertUser(user)
	}
}
