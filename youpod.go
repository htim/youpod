package youpod

import (
	"github.com/htim/youpod/auth"
	"io"
)

type User struct {
	Username   string `json:"username"`
	TelegramID int64  `json:"telegram_id"`

	GDriveToken auth.OAuth2Token `json:"g_drive_token"`

	FeedUrl string `json:"feed_url"`

	//list of file ids uploaded by user
	Files []string `json:"files"`
}

type UserService interface {
	SaveUser(u User) error
	FindUserByUsername(username string) (*User, error)
	FindUserByTelegramID(id int64) (*User, error)
	AddUserFile(u User, fileID string) error
}

type FileMetadata struct {
	ID          string
	Name        string
	ContentType string
	Length      int64
}

type File struct {
	FileMetadata
	Content io.ReadCloser
}

type FileService interface {
	SaveFile(f File, u User) (ID string, err error)
	GetFile(ID string, u User) (f *File, err error)
	GetFileMetadata(ID string) (m *FileMetadata, err error)
}

type YoutubeService interface {
	Download(owner User, link string) (File, error)
	Cleanup(File)
}
