package core

import (
	"context"
	"github.com/htim/youpod/auth"
)

type (
	User struct {
		Username   string `bson:"username"`
		TelegramID int64  `bson:"telegram_id"`

		GDriveToken auth.OAuth2Token `bson:"g_drive_token"`

		FeedUrl string `bson:"feed_url"`

		//list of file ids uploaded by user
		Files []string `bson:"files"`
	}

	UserRepository interface {
		SaveUser(ctx context.Context, u User) error
		FindUserByUsername(ctx context.Context, username string) (User, error)
		FindUserByTelegramID(ctx context.Context, id int64) (User, error)
		AddFileToUser(ctx context.Context, u User, fileID string) error
	}
)
