package core

import "github.com/htim/youpod/auth"

type (
	User struct {
		Username   string `json:"username"`
		TelegramID int64  `json:"telegram_id"`

		GDriveToken auth.OAuth2Token `json:"g_drive_token"`

		FeedUrl string `json:"feed_url"`

		//list of file ids uploaded by user
		Files []string `json:"files"`
	}

	UserRepository interface {
		SaveUser(u User) error
		FindUserByUsername(username string) (User, error)
		FindUserByTelegramID(id int64) (User, error)
		AddFileToUser(u User, fileID string) error
	}
)
