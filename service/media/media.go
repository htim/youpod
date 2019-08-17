package media

// Package media handles storing and retrieval of media files (typically audios extracted from YouTube videos)
// Implements core.MediaService

import (
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	"io"
)

type Store interface {
	GenerateID(user core.User) (ID string, err error)
	Save(user core.User, file core.File) (err error)
	Get(user core.User, ID string) (rs io.ReadSeeker, err error)
}

type Service struct {
	metadataService core.MetadataRepository
	store           Store
}

func NewService(metadataService core.MetadataRepository, store Store) *Service {
	return &Service{metadataService: metadataService, store: store}
}

func (s *Service) SaveFile(u core.User, f core.File) (string, error) {

	var err error
	if f.FileID == "" {
		f.FileID, err = s.store.GenerateID(u)
		if err != nil {
			return "", err
		}
	}

	if err := s.store.Save(u, f); err != nil {
		return "", err
	}

	if err := s.metadataService.SaveFileMetadata(u, f.Metadata); err != nil {
		return "", errors.Wrapf(err, "cannot save file metadata (user ID '%s', file ID '%s')", u.Username, f.FileID)
	}

	return f.FileID, nil
}

func (s *Service) GetFileContent(user core.User, fileID string) (io.ReadSeeker, error) {

	rs, err := s.store.Get(user, fileID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get file from store (user ID '%s', fileID '%s')", user.Username, fileID)
	}

	return rs, nil
}

func (s *Service) GetFileMetadata(user core.User, fileID string) (core.Metadata, error) {
	metadata, err := s.metadataService.GetFileMetadata(fileID)
	if err != nil {
		return core.Metadata{}, errors.Wrapf(err, "cannot load file metadata (user ID '%s', file ID '%s')", user.Username, fileID)
	}
	return metadata, nil
}
