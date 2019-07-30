package bolt

import (
	"encoding/json"
	"github.com/htim/youpod"
	gdrive "github.com/htim/youpod/google_drive"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

type folderMeta struct {
	OwnerID    string `json:"owner_id"`
	FolderName string `json:"folder_name"`
	FolderID   string `json:"folder_id"`
}

type FileService struct {
	client      *Client
	userService youpod.UserService
	drive       *gdrive.Client
	rootFolder  string
}

func NewFileService(client *Client, userService youpod.UserService, drive *gdrive.Client, rootFolder string) *FileService {
	return &FileService{client: client, userService: userService, drive: drive, rootFolder: rootFolder}
}

func (s *FileService) SaveFile(f youpod.File, u youpod.User) (string, error) {

	var folder folderMeta

	err := s.client.db.View(func(tx *bolt.Tx) error {
		foldersBkt := tx.Bucket(folders)
		if folderJson := foldersBkt.Get([]byte(u.Username)); folderJson != nil {
			if err := json.Unmarshal(folderJson, &folder); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return "", errors.Wrap(err, "cannot get folder meta")
	}

	if folder.FolderID == "" {
		folder.FolderID, err = s.drive.GenerateID(u)
		folder.FolderName = s.rootFolder
		if err != nil {
			return "", err
		}
		if err := s.drive.CreateRootFolder(u, folder.FolderName, folder.FolderID); err != nil {
			return "", err
		}
	} else {
		folderExists, err := s.drive.FolderExists(u, folder.FolderID)
		if err != nil {
			return "", err
		}
		if !folderExists {

			folder.FolderID, err = s.drive.GenerateID(u)
			if err != nil {
				return "", errors.Wrap(err, "cannot generate id for folder")
			}

			if err := s.drive.CreateRootFolder(u, folder.FolderName, folder.FolderID); err != nil {
				return "", err
			}
		}
	}

	err = s.client.db.Update(func(tx *bolt.Tx) error {
		foldersBkt := tx.Bucket(folders)
		folderJson, err := json.Marshal(folder)
		if err != nil {
			return err
		}
		if err = foldersBkt.Put([]byte(u.Username), folderJson); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", errors.Wrap(err, "cannot update folder meta")
	}

	id, err := s.drive.GenerateID(u)
	if err != nil {
		return "", errors.Wrap(err, "cannot generate id for file")
	}
	f.ID = id

	if err := s.drive.Put(u, f, folder.FolderID); err != nil {
		return "", errors.Wrap(err, "cannot save file in google drive")
	}

	err = s.client.db.Update(func(tx *bolt.Tx) error {
		fmJson, err := json.Marshal(f.FileMetadata)
		if err != nil {
			return err
		}

		bucket := tx.Bucket(filesBucket)
		if err = bucket.Put([]byte(f.ID), fmJson); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", errors.Wrap(err, "cannot save file meta")
	}

	return f.ID, nil
}

func (s *FileService) GetFile(ID string, u youpod.User) (f *youpod.File, err error) {

	var fileMeta youpod.FileMetadata

	err = s.client.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(filesBucket)
		fmJson := bkt.Get([]byte(ID))
		if fmJson == nil {
			return errNoValue
		}
		if err = json.Unmarshal(fmJson, &fileMeta); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if err == errNoValue {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "cannot get file metadata (file id: %s)", ID)
	}

	readCloser, err := s.drive.Get(u, fileMeta.ID)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get file from google drive")
	}

	return &youpod.File{
		FileMetadata: fileMeta,
		Content:      readCloser,
	}, nil

}

func (s *FileService) GetFileMetadata(ID string) (f *youpod.FileMetadata, err error) {

	var fm youpod.FileMetadata

	err = s.client.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(filesBucket)
		fmJson := bkt.Get([]byte(ID))

		if fmJson == nil {
			return errNoValue
		}

		if err := json.Unmarshal(fmJson, &fm); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		if err == errNoValue {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "cannot get file metadata (file id: %s)", ID)
	}

	return &fm, nil

}
