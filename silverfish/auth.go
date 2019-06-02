package silverfish

import (
	"time"
	"crypto/sha512"
	"encoding/hex"
	"strings"

	"github.com/jackey8616/Silverfish-backend/silverfish/entity"

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

func (a *Auth) sha512Str(src *string) *string {
	salted := strings.Join([]string{*src, *a.hashSalt}, "")
	h := sha512.New()
	h.Write([]byte(salted))
	s := hex.EncodeToString(h.Sum(nil))
	return &s
 }

// Register export
func (a *Auth) Register(account, password *string) *entity.APIResponse {
	hashedPassword := a.sha512Str(password)
	_, err := a.userInf.FindOne(bson.M{"account": *account}, &entity.User{})
	if err.Error() == "not found" {
		registerTime := time.Now()
		user := &entity.User{
			Account: *account,
			Password: *hashedPassword,
			RegisterDatetime: registerTime,
			LastLoginDatetime: registerTime,
		}
		a.userInf.Upsert(bson.M{"account": *account}, user)
		return &entity.APIResponse{
			Success: true,
			Data: &entity.User{
				Account: 		   user.Account,
				RegisterDatetime:  user.RegisterDatetime,
				LastLoginDatetime: user.LastLoginDatetime,
			},
		}
	} else if err != nil {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	return &entity.APIResponse{
		Fail: true,
		Data: map[string]string{"reason": "account exists."},
	}
}

// Login export
func (a *Auth) Login(account, password *string) *entity.APIResponse {
	hashedPassword := a.sha512Str(password)
	result, err := a.userInf.FindOne(bson.M{"account": *account}, &entity.User{})
	if err != nil {
		if err.Error() == "not found" {
			return &entity.APIResponse{
				Fail: true,
				Data: map[string]string{"reason": "Account not exists."},
			}
		}
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": err.Error()},
		}
	}
	user := result.(*entity.User)
	if user.Password != *hashedPassword {
		return &entity.APIResponse{
			Fail: true,
			Data: map[string]string{"reason": "Account or Password wrong."},
		}
	}
	apiResponse := &entity.APIResponse{
		Success: true,
		Data: &entity.User{
			Account: 		   user.Account,
			RegisterDatetime:  user.RegisterDatetime,
			LastLoginDatetime: user.LastLoginDatetime,
		},
	}
	user.LastLoginDatetime = time.Now()
	a.userInf.Upsert(bson.M{"account": account}, user)
	return apiResponse
}