package core

import "time"

type (
	Metadata struct {
		FileID      string
		TmpFileID   string
		Name        string
		ContentType string
		Author      string
		Size        int64  //size in bytes
		Picture     string //base64
		CreatedAt   time.Time
	}

	MetadataRepository interface {
		GetFileMetadata(ID string) (m Metadata, err error)
		SaveFileMetadata(u User, m Metadata) (err error)
	}
)
