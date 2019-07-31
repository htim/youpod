package media

import (
	"github.com/hashicorp/golang-lru"
	"github.com/htim/youpod"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

//FileSystemCache wraps stores and download requested file from store to filesystem to support byte-range requests
//Metadata is in in-memory LRU cache
type FileSystemCache struct {
	folder string
	cache  *lru.Cache
}

func NewFileSystemCache(folder string) (*FileSystemCache, error) {
	cache, err := lru.NewWithEvict(2, func(key interface{}, value interface{}) {
		k := key.(string)
		name := folder + "/" + k
		if err := os.Remove(name); err != nil {
			log.WithError(err).Warn("failed to remove evicted file")
		}
	})

	if err != nil {
		return nil, err
	}

	if err = os.RemoveAll(folder); err != nil {
		return nil, err
	}

	if err = os.Mkdir(folder, 0777); err != nil {
		return nil, err
	}

	return &FileSystemCache{folder: folder, cache: cache}, nil
}

func (s *FileSystemCache) GetFileContent(user youpod.User, store Store, fileID string) (*os.File, error) {
	name := s.folder + "/" + fileID

	if cached, ok := s.cache.Get(fileID); ok {
		return os.Open(cached.(string))
	}

	rc, err := store.Get(user, fileID)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get file from store (user ID '%s', fileID '%s', store '%s')", user.Username, fileID, store)
	}
	defer func() {
		if err := rc.Close(); err != nil {
			log.WithError(err).Error("cannot close store response")
		}
	}()

	file, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(file, rc); err != nil {
		return nil, err
	}

	s.cache.Add(fileID, name)

	return file, nil
}
