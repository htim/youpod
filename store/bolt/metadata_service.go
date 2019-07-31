package bolt

import (
	"github.com/htim/youpod"
	gdrive "github.com/htim/youpod/media/google_drive"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

type folderMeta struct {
	OwnerID    string `json:"owner_id"`
	FolderName string `json:"folder_name"`
	FolderID   string `json:"folder_id"`
}

type MetadataService struct {
	client      *Client
	userService youpod.UserService
	drive       *gdrive.Client
	rootFolder  string
}

func NewMetadataService(client *Client, userService youpod.UserService, drive *gdrive.Client, rootFolder string) *MetadataService {
	return &MetadataService{client: client, userService: userService, drive: drive, rootFolder: rootFolder}
}

func (s *MetadataService) SaveFile(f youpod.File, u youpod.User) (string, error) {

	//var folder folderMeta
	//
	//err := s.client.db.View(func(tx *bolt.Tx) error {
	//	foldersBkt := tx.Bucket(folders)
	//	if folderJson := foldersBkt.Get([]byte(u.Username)); folderJson != nil {
	//		if err := json.Unmarshal(folderJson, &folder); err != nil {
	//			return err
	//		}
	//	}
	//	return nil
	//})
	//
	//if err != nil {
	//	return "", errors.Wrap(err, "cannot get folder meta")
	//}
	//
	//if folder.FolderID == "" {
	//	folder.FolderID, err = s.drive.GenerateID(u)
	//	folder.FolderName = s.rootFolder
	//	if err != nil {
	//		return "", err
	//	}
	//	if err := s.drive.CreateRootFolder(u, folder.FolderName, folder.FolderID); err != nil {
	//		return "", err
	//	}
	//} else {
	//	folderExists, err := s.drive.FolderExists(u, folder.FolderID)
	//	if err != nil {
	//		return "", err
	//	}
	//	if !folderExists {
	//
	//		folder.FolderID, err = s.drive.GenerateID(u)
	//		if err != nil {
	//			return "", errors.Wrap(err, "cannot generate id for folder")
	//		}
	//
	//		if err := s.drive.CreateRootFolder(u, folder.FolderName, folder.FolderID); err != nil {
	//			return "", err
	//		}
	//	}
	//}
	//
	//err = s.client.db.Update(func(tx *bolt.Tx) error {
	//	foldersBkt := tx.Bucket(folders)
	//	folderJson, err := json.Marshal(folder)
	//	if err != nil {
	//		return err
	//	}
	//	if err = foldersBkt.Put([]byte(u.Username), folderJson); err != nil {
	//		return err
	//	}
	//	return nil
	//})
	//
	//if err != nil {
	//	return "", errors.Wrap(err, "cannot update folder meta")
	//}
	//
	//id, err := s.drive.GenerateID(u)
	//if err != nil {
	//	return "", errors.Wrap(err, "cannot generate id for file")
	//}
	//f.FileID = id
	//
	//if err := s.drive.Put(u, f, folder.FolderID); err != nil {
	//	return "", errors.Wrap(err, "cannot save file in google drive")
	//}
	//
	//err = s.client.db.Update(func(tx *bolt.Tx) error {
	//	fmJson, err := json.Marshal(f.FileMetadata)
	//	if err != nil {
	//		return err
	//	}
	//
	//	bucket := tx.Bucket(filesBucket)
	//	if err = bucket.Put([]byte(f.FileID), fmJson); err != nil {
	//		return err
	//	}
	//
	//	return nil
	//})
	//
	//if err != nil {
	//	return "", errors.Wrap(err, "cannot save file meta")
	//}
	//
	//return f.FileID, nil
	return "", nil
}

func (s *MetadataService) GetFileMetadata(ID string) (m youpod.FileMetadata, err error) {
	var fm youpod.FileMetadata

	err = s.client.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(filesBucket)
		if err := s.client.load(bkt, ID, &fm); err != nil {
			return errors.Wrapf(err, "failed to load key '%s' from bucket '%s'", ID, string(filesBucket))
		}
		return nil
	})

	if err != nil {
		if errors.Cause(err) == errNoValue {
			return youpod.FileMetadata{}, youpod.ErrMetadataNotFound
		}

		return youpod.FileMetadata{}, err
	}

	return fm, nil

}

func (s *MetadataService) SaveFileMetadata(u youpod.User, m youpod.FileMetadata) (err error) {

	if m.FileID == "" {
		return errors.New("FileID must be specified")
	}

	if m.StoreType == youpod.UnsetStore {
		return errors.New("StoreType must be specified")
	}

	err = s.client.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(filesBucket)
		if err = s.client.save(bucket, m.FileID, m); err != nil {
			return errors.Wrapf(err, "failed to save key '%s' to bucket '%s'", m.FileID, string(filesBucket))
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
