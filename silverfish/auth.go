package silverfish

import (
	"errors"
	"time"

	entity "silverfish/silverfish/entity"

	"go.mongodb.org/mongo-driver/bson"
)

// Auth export
type Auth struct {
	hashSalt    *string
	sessionSalt *string
	userInf     *entity.MongoInf
	sessionInf  *entity.MongoInf
}

// NewAuth export
func NewAuth(hashSalt *string, userInf, sessionInf *entity.MongoInf) *Auth {
	saltTmp := "SILVERFISH"
	a := new(Auth)
	a.hashSalt = hashSalt
	a.userInf = userInf
	a.sessionInf = sessionInf
	a.sessionSalt = &saltTmp
	return a
}

func (a *Auth) findSession(token *string) (*entity.Session, error) {
	result, err := a.sessionInf.FindOne(bson.M{"token": *token}, &entity.Session{})
	if err != nil {
		return nil, err
	}
	return result.(*entity.Session), nil
}

// GetSession export
func (a *Auth) GetSession(sessionToken *string) (*entity.Session, error) {
	session, err := a.findSession(sessionToken)
	if err != nil {
		return nil, errors.New("SessionToken not exists")
	}
	session.KeepAlive()
	a.sessionInf.Update(bson.M{"token": *sessionToken}, session)
	return session, nil
}

// InsertSession export
func (a *Auth) InsertSession(user *entity.User, keepLogin bool) *entity.Session {
	payload := user.Account + time.Now().String()
	sessionToken := SHA512Str(&payload, a.sessionSalt)
	session := entity.NewSession(keepLogin, sessionToken, user)
	a.sessionInf.Insert(session)
	return session
}

// IsTokenValid export
func (a *Auth) IsTokenValid(sessionToken *string) bool {
	session, err := a.findSession(sessionToken)
	if err != nil {
		return false
	}
	if session.IsExpired() {
		a.sessionInf.Remove(bson.M{"token": *sessionToken})
		return false
	}
	session.KeepAlive()
	a.sessionInf.Update(bson.M{"token": *sessionToken}, session)
	return true
}

// KillSession export
func (a *Auth) KillSession(sessionToken *string) bool {
	err := a.sessionInf.Remove(bson.M{"token": *sessionToken})
	return err == nil
}

// Register export
func (a *Auth) Register(isAdmin bool, account, password *string) (*entity.User, error) {
	hashedPassword := SHA512Str(password, a.hashSalt)
	_, err := a.userInf.FindOne(bson.M{"account": *account}, &entity.User{})
	if err != nil {
		if err.Error() == "not found" {
			registerTime := time.Now()
			user := &entity.User{
				IsAdmin:           isAdmin,
				Account:           *account,
				Password:          *hashedPassword,
				RegisterDatetime:  registerTime,
				LastLoginDatetime: registerTime,
				Bookmark:          &entity.Bookmark{},
			}
			a.userInf.Upsert(bson.M{"account": *account}, user)
			return &entity.User{
				IsAdmin:           user.IsAdmin,
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
		IsAdmin:           user.IsAdmin,
		Account:           user.Account,
		RegisterDatetime:  user.RegisterDatetime,
		LastLoginDatetime: user.LastLoginDatetime,
		Bookmark:          user.Bookmark,
	}, nil
}

// IsAdmin export
func (a *Auth) IsAdmin(account *string) (bool, error) {
	result, err := a.userInf.FindOne(bson.M{"account": account}, &entity.User{})
	if err != nil {
		return false, errors.New("Account not exists")
	}
	return result.(*entity.User).IsAdmin, nil
}
