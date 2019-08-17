package bolt

import (
	"github.com/htim/youpod"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

type metadataRepository struct {
	client         *Client
	userRepository core.UserRepository
	rootFolder     string
}

func NewMetadataRepository(client *Client, userService core.UserRepository, rootFolder string) core.MetadataRepository {
	return &metadataRepository{client: client, userRepository: userService, rootFolder: rootFolder}
}

func (r *metadataRepository) GetFileMetadata(ID string) (m core.Metadata, err error) {
	var fm core.Metadata

	err = r.client.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(filesBucket)
		if err := r.client.load(bkt, ID, &fm); err != nil {
			return errors.Wrapf(err, "failed to load key '%s' from bucket '%s'", ID, string(filesBucket))
		}
		return nil
	})

	if err != nil {
		if errors.Cause(err) == errNoValue {
			return core.Metadata{}, youpod.ErrMetadataNotFound
		}

		return core.Metadata{}, err
	}

	return fm, nil

}

func (r *metadataRepository) SaveFileMetadata(u core.User, m core.Metadata) (err error) {

	if m.FileID == "" {
		return errors.New("FileID must be specified")
	}

	err = r.client.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(filesBucket)
		if err = r.client.save(bucket, m.FileID, m); err != nil {
			return errors.Wrapf(err, "failed to save key '%s' to bucket '%s'", m.FileID, string(filesBucket))
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
