package core

import "io"

type (
	File struct {
		Metadata
		Content io.ReadCloser
	}
)
