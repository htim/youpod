package youpod

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrFileNotFound     = errors.New("file not found")
	ErrMetadataNotFound = errors.New("file metadata not found")
)
