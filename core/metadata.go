package core

import (
	"context"
	"time"
)

type (
	Metadata struct {
		FileID      string    `bson:"file_id"`
		TmpFileID   string    `bson:"tmp_file_id"`
		Name        string    `bson:"name"`
		ContentType string    `bson:"content_type"`
		Author      string    `bson:"author"`
		Size        int64     `bson:"size"`    //size in bytes
		Picture     string    `bson:"picture"` //base64
		CreatedAt   time.Time `bson:"created_at"`
	}

	MetadataRepository interface {
		GetFileMetadata(ctx context.Context, ID string) (m Metadata, err error)
		SaveFileMetadata(ctx context.Context, m Metadata) (err error)
	}
)
