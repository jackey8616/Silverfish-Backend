package entity

import "time"

// Session export
type Session struct {
	keepLogin bool
	token     *string
	account   *string
	loginTS   time.Time
	expireTS  time.Time
}

// NewSession export
func NewSession(keepLogin bool, token *string, user *User) *Session {
	s := new(Session)
	s.keepLogin = keepLogin
	s.token = token
	s.account = &(*user).Account
	s.loginTS = time.Now()
	if keepLogin == true {
		s.expireTS = time.Now().Add(time.Hour * 24 * 7)
	} else {
		s.expireTS = time.Now().Add(time.Hour)
	}
	return s
}

// GetAccount export
func (s *Session) GetAccount() *string {
	return s.account
}

// KeepAlive export
func (s *Session) KeepAlive() {
	if s.keepLogin == true {
		s.expireTS = time.Now().Add(time.Hour * 24 * 7)
	} else {
		s.expireTS = time.Now().Add(time.Hour)
	}
}

// IsExpired export
func (s *Session) IsExpired() bool {
	return time.Now().After(s.expireTS)
}
