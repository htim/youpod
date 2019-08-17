package core

import "io"

type (
	MediaService interface {
		SaveFile(u User, f File) (string, error)
		GetFileContent(u User, fileID string) (io.ReadSeeker, error)
		GetFileMetadata(user User, fileID string) (Metadata, error)
	}
)
