package silverfish

import (
	"errors"
	"time"

	entity "silverfish/silverfish/entity"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
)

// Auth export
type Auth struct {
	hashSalt    *string
	sessionSalt *string
	userInf     *entity.DynamoInf
	sessions    map[string]*entity.Session
}

// NewAuth export
func NewAuth(hashSalt *string, userInf *entity.DynamoInf) *Auth {
	saltTmp := "SILVERFISH"
	a := new(Auth)
	a.hashSalt = hashSalt
	a.userInf = userInf
	a.sessionSalt = &saltTmp
	a.sessions = map[string]*entity.Session{}
	return a
}

// ExpireLoop export
func (a *Auth) ExpireLoop() {
	newSessions := map[string]*entity.Session{}
	for k, v := range a.sessions {
		if v.IsExpired() == false {
			newSessions[k] = v
		}
	}
	a.sessions = newSessions
}

// GetSession export
func (a *Auth) GetSession(sessionToken *string) (*entity.Session, error) {
	if _, ok := a.sessions[*sessionToken]; ok {
		session := a.sessions[*sessionToken]
		session.KeepAlive()
		return session, nil
	}
	return nil, errors.New("SessionToken not exists")
}

// InsertSession export
func (a *Auth) InsertSession(user *entity.User, keepLogin bool) *entity.Session {
	a.ExpireLoop()
	payload := (*user).Account + time.Now().String()
	sessionToken := SHA512Str(&payload, a.sessionSalt)
	if _, ok := a.sessions[*sessionToken]; ok {
		return a.InsertSession(user, keepLogin)
	}
	session := entity.NewSession(keepLogin, sessionToken, user)
	a.sessions[*sessionToken] = session
	return session
}

// IsTokenValid export
func (a *Auth) IsTokenValid(sessionToken *string) bool {
	if val, ok := a.sessions[*sessionToken]; ok {
		result := val.IsExpired()
		if result == true {
			delete(a.sessions, *sessionToken)
			return false
		}
		a.sessions[*sessionToken].KeepAlive()
		return true
	}
	return false
}

// KillSession export
func (a *Auth) KillSession(sessionToken *string) bool {
	if _, ok := a.sessions[*sessionToken]; ok {
		delete(a.sessions, *sessionToken)
		return true
	}
	return false
}

func (a *Auth) findByAccount(account string) (*entity.User, error) {
	keyCond := expression.Key("account").Equal(expression.Value(account))

	result, err := a.userInf.FindOne(&keyCond, &entity.User{})
	if err != nil {
		return nil, err
	}

	return result.(*entity.User), nil
}

func (a *Auth) upsertUser(user *entity.User) (*entity.User, error) {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, err
	}

	_, putErr := a.userInf.Upsert(item)
	return user, putErr
}

// Register export
func (a *Auth) Register(isAdmin bool, account, password *string) (*entity.User, error) {
	hashedPassword := SHA512Str(password, a.hashSalt)
	_, err := a.findByAccount(*account)

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
			a.upsertUser(user)
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
	user, err := a.findByAccount(*account)
	if err != nil {
		if err.Error() == "not found" {
			return nil, errors.New("Account not exists")
		}
		return nil, err
	}
	if user.Password != *hashedPassword {
		return nil, errors.New("Account or Password wrong")
	}

	user.LastLoginDatetime = time.Now()
	a.upsertUser(user)
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
	user, err := a.findByAccount(*account)
	if err != nil {
		return false, errors.New("Account not exists")
	}
	return user.IsAdmin, nil
}
