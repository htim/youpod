package core

type (
	Metadata struct {
		FileID      string
		TmpFileID   string
		Name        string
		ContentType string
		Length      int64
		Picture     string //base64
	}

	MetadataRepository interface {
		GetFileMetadata(ID string) (m Metadata, err error)
		SaveFileMetadata(u User, m Metadata) (err error)
	}
)
