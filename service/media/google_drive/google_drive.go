package gdrive

import (
	"context"
	"fmt"
	"github.com/htim/youpod/auth"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
)

type Client struct {
	userRepository core.UserRepository
	config         oauth2.Config
}

func NewClient(
	userRepository core.UserRepository,
	clientID, clientSecret, redirectUrl string,
) *Client {
	return &Client{
		userRepository: userRepository,
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

func (c *Client) Save(user core.User, file core.File) (err error) {

	filesService, err := c.filesService(user)
	if err != nil {
		return errors.Wrap(err, "cannot init google drive api client")
	}

	if file.FileID == "" {
		return errors.New("file id must be specified")
	}

	driveFile := &drive.File{
		Id:   file.FileID,
		Name: file.Name,
	}

	_, err = filesService.Create(driveFile).Media(file.Content).Do()
	if err != nil {
		return errors.Wrap(err, "cannot upload file")
	}

	return nil
}

func (c *Client) Get(user core.User, ID string) (io.ReadSeeker, error) {
	filesService, err := c.filesService(user)
	if err != nil {
		return nil, errors.Wrap(err, "cannot init google drive api client")
	}

	file, err := filesService.Get(ID).Fields("size").Do()
	if err != nil {
		return nil, errors.Wrap(err, "cannot load file from google drive")
	}

	return &readSeeker{
		user:     user,
		client:   c,
		fileID:   ID,
		fileSize: file.Size,
		offset:   0,
	}, nil

}

func (c *Client) FolderExists(user core.User, folderID string) (bool, error) {
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

func (c *Client) CreateRootFolder(user core.User, folderName, folderID string) error {
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

func (c *Client) GenerateID(user core.User) (string, error) {
	filesService, err := c.filesService(user)
	if err != nil {
		return "", errors.Wrap(err, "cannot init google drive api")
	}
	generatedID, err := filesService.GenerateIds().Count(1).Do()
	if err != nil {
		return "", errors.Wrap(err, "cannot generate google drive FileID")
	}
	return generatedID.Ids[0], nil
}

func (c *Client) filesService(user core.User) (*drive.FilesService, error) {

	if user.GDriveToken.IsExpired() {
		newToken, err := c.RefreshToken(user.GDriveToken)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot refresh google drive token for user: %s", user.Username)
		}
		user.GDriveToken = newToken
		if err = c.userRepository.SaveUser(context.Background(), user); err != nil {
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

type readSeeker struct {
	user     core.User
	client   *Client
	fileID   string
	fileSize int64
	offset   int64
}

func (s *readSeeker) Read(p []byte) (int, error) {
	if s.offset >= s.fileSize {
		return 0, io.EOF
	}

	filesService, err := s.client.filesService(s.user)
	if err != nil {
		return 0, err
	}

	bytesHeader := fmt.Sprintf("bytes=%d-%d", s.offset, s.offset+int64(len(p)-1))
	getCall := filesService.Get(s.fileID)
	getCall.Header().Set("Range", bytesHeader)
	response, err := getCall.Download()
	if err != nil {
		return 0, err
	}
	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	n := copy(p, buf)
	s.offset += int64(n)
	return len(p), nil
}

func (s *readSeeker) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = s.offset + offset
	case io.SeekEnd:
		abs = s.fileSize + offset
	default:
		return 0, errors.New("gdrive.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("gdrive.Reader.Seek: negative position")
	}
	s.offset = abs
	return abs, nil
}
