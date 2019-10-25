package silverfish

import (
	"errors"
	entity "silverfish/silverfish/entity"
	"time"
)

// SessionUsecase export
type SessionUsecase struct {
	sessionSalt *string
	sessions    map[string]*entity.Session
}

// NewSessionUsecase export
func NewSessionUsecase() *SessionUsecase {
	saltTmp := "SILVERFISH"
	su := new(SessionUsecase)
	su.sessionSalt = &saltTmp
	su.sessions = map[string]*entity.Session{}
	return su
}

// ExpireLoop export
func (su *SessionUsecase) ExpireLoop() {
	newSessions := map[string]*entity.Session{}
	for k, v := range su.sessions {
		if v.IsExpired() == false {
			newSessions[k] = v
		}
	}
	su.sessions = newSessions
}

// GetSession export
func (su *SessionUsecase) GetSession(sessionToken *string) (*entity.Session, error) {
	if _, ok := su.sessions[*sessionToken]; ok {
		session := su.sessions[*sessionToken]
		session.KeepAlive()
		return session, nil
	}
	return nil, errors.New("SessionToken not exists")
}

// InsertSession export
func (su *SessionUsecase) InsertSession(user *entity.User, keepLogin bool) *string {
	su.ExpireLoop()
	payload := (*user).Account + time.Now().String()
	sessionToken := SHA512Str(&payload, su.sessionSalt)
	if _, ok := su.sessions[*sessionToken]; ok {
		return su.InsertSession(user, keepLogin)
	}
	su.sessions[*sessionToken] = entity.NewSession(keepLogin, sessionToken, user)
	return sessionToken
}

// IsTokenValid export
func (su *SessionUsecase) IsTokenValid(sessionToken *string) bool {
	if val, ok := su.sessions[*sessionToken]; ok {
		result := val.IsExpired()
		if result == true {
			delete(su.sessions, *sessionToken)
			return false
		}
		su.sessions[*sessionToken].KeepAlive()
		return true
	}
	return false
}

// KillSession export
func (su *SessionUsecase) KillSession(sessionToken *string) bool {
	if _, ok := su.sessions[*sessionToken]; ok {
		delete(su.sessions, *sessionToken)
		return true
	}
	return false
}
