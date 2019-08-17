package auth

import (
	"time"
)

type OAuth2 interface {
	URL(state string) string
	Exchange(code string) (OAuth2Token, error)
	RefreshToken(t OAuth2Token) (OAuth2Token, error)
	GetUserInfo(t OAuth2Token) (UserInfo, error)
}

type OAuth2Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func (t OAuth2Token) IsExpired() bool {
	return time.Now().After(t.Expiry)
}

type UserInfo struct {
	Email       string
	DisplayName string
}

type OAuth2TokenSource interface {
	Token() (OAuth2Token, error)
}
