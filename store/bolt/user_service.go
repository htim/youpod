package bolt

import (
	"encoding/binary"
	"encoding/json"
	"github.com/htim/youpod"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	tgChatIdBucket = []byte("tgChatId")
)

type UserService struct {
	client *Client
}

func NewUserService(client *Client) (*UserService, error) {

	if err := client.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(userBucket)
		if _, err := bkt.CreateBucketIfNotExists(tgChatIdBucket); err != nil {
			return errors.Wrap(err, "cannot create tg bucket for users")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "cannot set up users buckets")
	}

	return &UserService{
		client: client,
	}, nil
}

func (s *UserService) SaveUser(u youpod.User) error {
	jval, err := json.Marshal(u)
	if err != nil {
		return errors.Wrap(err, "cannot marshal user for saving")
	}
	err = s.client.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(userBucket)
		if err := bkt.Put([]byte(u.Username), jval); err != nil {
			return errors.Wrap(err, "cannot save user into root bucket")
		}
		tgBkt := bkt.Bucket(tgChatIdBucket)
		buf := make([]byte, binary.MaxVarintLen64)
		binary.PutVarint(buf, u.TelegramID)
		if err := tgBkt.Put(buf, []byte(u.Username)); err != nil {
			return errors.Wrap(err, "cannot save tg chat id key")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "cannot save user")
	}

	return nil
}

func (s *UserService) FindUserByUsername(username string) (*youpod.User, error) {
	var u youpod.User
	err := s.client.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(userBucket)
		jval := bkt.Get([]byte(username))
		if jval == nil {
			return errNoValue
		}
		if err := json.Unmarshal(jval, &u); err != nil {
			return errors.Wrap(err, "cannot unmarshal stored user")
		}
		return nil
	})

	if err != nil {

		if err == errNoValue {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "cannot get user: %s", username)
	}

	return &u, nil
}

func (s *UserService) FindUserByTelegramID(id int64) (*youpod.User, error) {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, id)

	var u youpod.User

	err := s.client.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(userBucket)
		tgBkt := bkt.Bucket(tgChatIdBucket)
		username := tgBkt.Get(buf)
		if username == nil {
			return errNoValue
		}
		user := bkt.Get(username)
		if user == nil {
			return errNoValue
		}
		if err := json.Unmarshal(user, &u); err != nil {
			return errors.Wrap(err, "cannot unmarshal stored user")
		}
		return nil
	})

	if err != nil {

		if err == errNoValue {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "cannot get user: %d", id)
	}

	return &u, nil
}

func (s *UserService) AddUserFile(u youpod.User, fileID string) error {
	user, err := s.FindUserByUsername(u.Username)
	if err != nil {
		return errors.Wrap(err, "cannot find user")
	}
	if user.Files == nil {
		user.Files = make([]string, 0)
	}
	user.Files = append(user.Files, fileID)
	if err = s.SaveUser(*user); err != nil {
		return errors.Wrap(err, "cannot update user")
	}
	return nil
}
