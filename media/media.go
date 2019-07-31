// Package media handles storing and retrieval of media files (typically audios extracted from YouTube videos)
// Provides Store with Save and Load
// Service object encloses Store and add common methods, this is the one consumer should use
package media

import (
	"github.com/htim/youpod"
	"github.com/pkg/errors"
	"io"
)

type Store interface {
	GenerateID(user youpod.User) (ID string, err error)
	Save(user youpod.User, file youpod.File) (err error)
	Get(user youpod.User, ID string) (rs io.ReadSeeker, err error)
}

type Service struct {
	metadataService youpod.MetadataService
	stores          map[youpod.StoreType]Store
}

func NewService(metadataService youpod.MetadataService, stores map[youpod.StoreType]Store) *Service {
	return &Service{metadataService: metadataService, stores: stores}
}

func (s *Service) SaveFile(u youpod.User, f youpod.File) (string, error) {

	store, err := s.getStore(u.DefaultStoreType)
	if err != nil {
		return "", err
	}

	if f.FileID == "" {
		f.FileID, err = store.GenerateID(u)
		if err != nil {
			return "", err
		}
	}

	if f.StoreType == youpod.UnsetStore {
		f.StoreType = u.DefaultStoreType
	}

	if err := store.Save(u, f); err != nil {
		return "", err
	}

	if err := s.metadataService.SaveFileMetadata(u, f.FileMetadata); err != nil {
		return "", errors.Wrapf(err, "cannot save file metadata (user ID '%s', file ID '%s')", u.Username, f.FileID)
	}

	return f.FileID, nil
}

func (s *Service) GetFileContent(user youpod.User, fileID string) (io.ReadSeeker, error) {

	metadata, err := s.metadataService.GetFileMetadata(fileID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load file metadata (user ID '%s', file ID '%s')", user.Username, fileID)
	}

	store, err := s.getStore(metadata.StoreType)
	if err != nil {
		return nil, err
	}

	rs, err := store.Get(user, fileID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get file from store (user ID '%s', fileID '%s', store '%s'", user.Username, fileID, metadata.StoreType)
	}

	return rs, nil
}

func (s *Service) GetFileMetadata(user youpod.User, fileID string) (youpod.FileMetadata, error) {
	metadata, err := s.metadataService.GetFileMetadata(fileID)
	if err != nil {
		return youpod.FileMetadata{}, errors.Wrapf(err, "cannot load file metadata (user ID '%s', file ID '%s')", user.Username, fileID)
	}
	return metadata, nil
}

func (s *Service) getStore(t youpod.StoreType) (Store, error) {
	store, ok := s.stores[t]
	if !ok {
		return nil, errors.Errorf("unregistered store type: %s", t)
	}
	return store, nil
}
