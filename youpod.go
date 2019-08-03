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

	DefaultStoreType StoreType
}

type UserService interface {
	SaveUser(u User) error
	FindUserByUsername(username string) (User, error)
	FindUserByTelegramID(id int64) (User, error)
	AddUserFile(u User, fileID string) error
}

type FileMetadata struct {
	FileID      string
	TmpFileID   string
	Name        string
	ContentType string
	Length      int64
	Picture     string //base64
	StoreType   StoreType
}

type File struct {
	FileMetadata
	Content io.ReadCloser
}

type MetadataService interface {
	GetFileMetadata(ID string) (m FileMetadata, err error)
	SaveFileMetadata(u User, m FileMetadata) (err error)
}

type YoutubeService interface {
	Download(owner User, link string) (File, error)
	Cleanup(File)
}

type StoreType int

const (
	UnsetStore StoreType = iota
	GoogleDrive
	Dropbox
)

func (s StoreType) String() string {
	switch s {
	case Dropbox:
		return "Dropbox"
	case GoogleDrive:
		return "Google Drive"
	default:
		return "Unknown store"
	}
}
