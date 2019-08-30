package core

import (
	"context"
	"io"
)

type (
	MediaService interface {
		SaveFile(u User, f File) (string, error)
		GetFileContent(u User, fileID string) (io.ReadSeeker, error)
		GetFileMetadata(user User, fileID string, ctx context.Context) (Metadata, error)
	}
)
