package entity

import "time"

// Session export
type Session struct {
	KeepLogin bool      `json:"keepLogin" bson:"keepLogin"`
	Token     string    `json:"token" bson:"token"`
	Account   string    `json:"account" bson:"account"`
	LoginTS   time.Time `json:"loginTS" bson:"loginTS"`
	ExpireTS  time.Time `json:"expireTS" bson:"expireTS"`
}

// NewSession export
func NewSession(keepLogin bool, token *string, user *User) *Session {
	s := &Session{
		KeepLogin: keepLogin,
		Token:     *token,
		Account:   user.Account,
		LoginTS:   time.Now(),
	}
	s.refreshExpiry()
	return s
}

func (s *Session) refreshExpiry() {
	if s.KeepLogin {
		s.ExpireTS = time.Now().Add(time.Hour * 24 * 7)
	} else {
		s.ExpireTS = time.Now().Add(time.Hour)
	}
}

// GetToken export
func (s *Session) GetToken() *string { return &s.Token }

// GetExpireTS export
func (s *Session) GetExpireTS() time.Time { return s.ExpireTS }

// GetAccount export
func (s *Session) GetAccount() *string { return &s.Account }

// KeepAlive export
func (s *Session) KeepAlive() { s.refreshExpiry() }

// IsExpired export
func (s *Session) IsExpired() bool { return time.Now().After(s.ExpireTS) }
