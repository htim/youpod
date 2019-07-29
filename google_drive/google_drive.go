package gdrive

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/htim/youpod"
	"github.com/htim/youpod/auth"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

//gdrive.Client implements both FilesService and auth.OAuth2 to provide token refreshing
type Client struct {
	userService youpod.UserService
	config      oauth2.Config
}

func NewClient(
	userService youpod.UserService,
	clientID, clientSecret, redirectUrl string,
) *Client {
	return &Client{
		userService: userService,
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectUrl,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"https://www.googleapis.com/auth/drive.file",
				"https://www.googleapis.com/auth/drive.metadata",
			},
		},
	}
}

func (c *Client) Put(user youpod.User, f youpod.File, root string) error {

	filesService, err := c.filesService(user)
	if err != nil {
		return errors.Wrap(err, "cannot init google drive api client")
	}

	if f.ID == "" {
		return errors.New("file id must be specified")
	}

	driveFile := &drive.File{
		Id:      f.ID,
		Name:    f.Name,
		Parents: []string{root},
	}

	_, err = filesService.Create(driveFile).Media(f.Content).Do()
	if err != nil {
		return errors.Wrap(err, "cannot upload file")
	}

	return nil
}

func (c *Client) Get(owner youpod.User, id string) (io.ReadCloser, error) {
	return nil, nil
}

func (c *Client) FolderExists(user youpod.User, folderID string) (bool, error) {
	filesService, err := c.filesService(user)

	if err != nil {
		return false, errors.Wrap(err, "cannot init google drive api client")
	}

	f, err := filesService.Get(folderID).Do()
	if err != nil {
		if err2, ok := err.(*googleapi.Error); ok {
			if err2.Code == 404 {
				return false, nil
			}
		}
		return false, errors.Wrap(err, "cannot check folder existence")
	}

	fmt.Println(f)

	return true, nil
}

func (c *Client) CreateRootFolder(user youpod.User, folderName, folderID string) error {
	filesService, err := c.filesService(user)

	if err != nil {
		return errors.Wrap(err, "cannot init google drive api client")
	}

	_, err = filesService.Get(folderID).Do()

	if err != nil {

		if err2, ok := err.(*googleapi.Error); ok {

			if err2.Code == 404 {
				createCall := filesService.Create(&drive.File{
					Id:       folderID,
					MimeType: "application/vnd.google-apps.folder",
					Name:     "YouPod",
				})
				if _, err := createCall.Do(); err != nil {
					return errors.Wrap(err, "cannot create folder")
				}
				return nil
			}

		}

		return errors.Wrap(err, "cannot get folder")
	}

	return nil
}

func (c *Client) GenerateID(user youpod.User) (string, error) {
	filesService, err := c.filesService(user)
	if err != nil {
		return "", errors.Wrap(err, "cannot init google drive api")
	}
	generatedID, err := filesService.GenerateIds().Count(1).Do()
	if err != nil {
		return "", errors.Wrap(err, "cannot generate google drive ID")
	}
	return generatedID.Ids[0], nil
}

func (c *Client) filesService(user youpod.User) (*drive.FilesService, error) {

	if user.GDriveToken.IsExpired() {
		newToken, err := c.RefreshToken(user.GDriveToken)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot refresh google drive token for user: %s", user.Username)
		}
		user.GDriveToken = newToken
		if err = c.userService.SaveUser(user); err != nil {
			return nil, errors.Wrapf(err, "cannot update google drive token for user: %s", user.Username)
		}
	}

	ts, err := tokenSource(user.GDriveToken)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert token source")
	}

	service, err := drive.NewService(context.Background(), option.WithTokenSource(ts))
	if err != nil {
		return nil, errors.Wrap(err, "cannot init google drive client")
	}

	return drive.NewFilesService(service), nil
}

func tokenSource(token auth.OAuth2Token) (oauth2.TokenSource, error) {
	tok := oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	return oauth2.StaticTokenSource(&tok), nil
}

func (c *Client) URL(state string) string {
	return c.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (c *Client) Exchange(code string) (auth.OAuth2Token, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	token, err := c.config.Exchange(ctx, code)
	if err != nil {
		return auth.OAuth2Token{}, errors.Wrap(err, "cannot exchange code to token for google drive")
	}

	fmt.Println(token.Expiry.String())

	return auth.OAuth2Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}, nil
}

func (c *Client) RefreshToken(t auth.OAuth2Token) (auth.OAuth2Token, error) {

	src := oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		Expiry:       t.Expiry,
	}

	tokenSource := c.config.TokenSource(context.Background(), &src)

	newToken, err := tokenSource.Token()
	if err != nil {
		return auth.OAuth2Token{}, errors.Wrap(err, "cannot obtain new token")
	}

	return auth.OAuth2Token{
		AccessToken:  newToken.AccessToken,
		RefreshToken: newToken.RefreshToken,
		Expiry:       newToken.Expiry,
	}, nil
}

func (c *Client) GetUserInfo(t auth.OAuth2Token) (auth.UserInfo, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	baseUrl := "https://www.googleapis.com/drive/v3/about"
	req, err := http.NewRequest(http.MethodGet, baseUrl, nil)
	if err != nil {
		return auth.UserInfo{}, errors.Wrapf(err, "cannot construct new request: %s", baseUrl)
	}

	req = req.WithContext(ctx)

	req.Header.Add("Authorization", "Bearer "+t.AccessToken)

	query := req.URL.Query()
	query.Add("fields", "user")
	req.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return auth.UserInfo{}, errors.Wrapf(err, "cannot make request: %s", baseUrl)
	}
	defer resp.Body.Close()

	user, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return auth.UserInfo{}, errors.Wrapf(err, "cannot read response: %s", baseUrl)
	}

	var userShape struct {
		User struct {
			DisplayName  string `json:"displayName"`
			EmailAddress string `json:"emailAddress"`
		} `json:"user"`
	}

	if err := json.Unmarshal(user, &userShape); err != nil {
		return auth.UserInfo{}, errors.Wrapf(err, "cannot unmarshal response: %s", baseUrl)
	}

	return auth.UserInfo{
		Email:       userShape.User.EmailAddress,
		DisplayName: userShape.User.DisplayName,
	}, nil

}
