package gdrive

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/htim/youpod/auth"
//	"github.com/pkg/errors"
//	"golang.org/x/oauth2"
//	"golang.org/x/oauth2/google"
//	"io/ioutil"
//	"net/http"
//	"time"
//)
//
//type Auth struct {
//	config oauth2.Config
//}
//
//func NewAuth(clientID, clientSecret, redirectUrl string) *Auth {
//	return &Auth{
//		config: oauth2.Config{
//			ClientID:     clientID,
//			ClientSecret: clientSecret,
//			RedirectURL:  redirectUrl,
//			Endpoint:     google.Endpoint,
//			Scopes: []string{
//				"https://www.googleapis.com/auth/drive.file",
//				"https://www.googleapis.com/auth/drive.metadata",
//			},
//		},
//	}
//}
//
//func (a *Auth) URL(state string) string {
//	return a.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
//}
//
//func (a *Auth) Exchange(code string) (auth.OAuth2Token, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	token, err := a.config.Exchange(ctx, code)
//	if err != nil {
//		return auth.OAuth2Token{}, errors.Wrap(err, "cannot exchange code to token for google drive")
//	}
//
//	fmt.Println(token.Expiry.String())
//
//	return auth.OAuth2Token{
//		AccessToken:  token.AccessToken,
//		RefreshToken: token.RefreshToken,
//		Expiry:       token.Expiry,
//	}, nil
//}
//
//func (a *Auth) RefreshToken(t auth.OAuth2Token) (auth.OAuth2Token, error) {
//
//	src := oauth2.Token{
//		AccessToken:  t.AccessToken,
//		RefreshToken: t.RefreshToken,
//		Expiry:       t.Expiry,
//	}
//
//	tokenSource := a.config.TokenSource(context.Background(), &src)
//
//	newToken, err := tokenSource.Token()
//	if err != nil {
//		return auth.OAuth2Token{}, errors.Wrap(err, "cannot obtain new token")
//	}
//
//	return auth.OAuth2Token{
//		AccessToken:  newToken.AccessToken,
//		RefreshToken: newToken.RefreshToken,
//		Expiry:       newToken.Expiry,
//	}, nil
//}
//
//func (a *Auth) GetUserInfo(t auth.OAuth2Token) (auth.UserInfo, error) {
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	baseUrl := "https://www.googleapis.com/drive/v3/about"
//	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
//	if err != nil {
//		return auth.UserInfo{}, errors.Wrapf(err, "cannot construct new request: %s", baseUrl)
//	}
//
//	req = req.WithContext(ctx)
//
//	req.Header.Add("Authorization", "Bearer "+t.AccessToken)
//
//	query := req.URL.Query()
//	query.Add("fields", "user")
//	req.URL.RawQuery = query.Encode()
//
//	resp, err := http.DefaultClient.Do(req)
//	if err != nil {
//		return auth.UserInfo{}, errors.Wrapf(err, "cannot make request: %s", baseUrl)
//	}
//	defer resp.Body.Close()
//
//	user, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return auth.UserInfo{}, errors.Wrapf(err, "cannot read response: %s", baseUrl)
//	}
//
//	var userShape struct {
//		User struct {
//			DisplayName  string `json:"displayName"`
//			EmailAddress string `json:"emailAddress"`
//		} `json:"user"`
//	}
//
//	if err := json.Unmarshal(user, &userShape); err != nil {
//		return auth.UserInfo{}, errors.Wrapf(err, "cannot unmarshal response: %s", baseUrl)
//	}
//
//	return auth.UserInfo{
//		Email:       userShape.User.EmailAddress,
//		DisplayName: userShape.User.DisplayName,
//	}, nil
//
//}
//
//type tokenSource struct {
//	auth      *Auth
//	token     auth.OAuth2Token
//	onRefresh func(token auth.OAuth2Token) error
//}
//
//func (src *tokenSource) Token() (auth.OAuth2Token, error) {
//	if src.token.IsExpired() {
//		newToken, err := src.auth.RefreshToken(src.token)
//		if err != nil {
//			return auth.OAuth2Token{}, errors.Wrap(err, "cannot refresh token")
//		}
//		if err = src.onRefresh(newToken); err != nil {
//			return auth.OAuth2Token{}, errors.Wrap(err, "cannot call refresh callback")
//		}
//		return newToken, nil
//	}
//	return src.token, nil
//}
//
//func (a *Auth) TokenSource(t auth.OAuth2Token, onRefresh func(newToken auth.OAuth2Token) error) auth.OAuth2TokenSource {
//	return &tokenSource{
//		auth:      a,
//		token:     t,
//		onRefresh: onRefresh,
//	}
//}
