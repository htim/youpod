package bolt

import (
	"context"
	"github.com/htim/youpod"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"strconv"
)

var (
	tgChatIdBucket = []byte("tgChatId")
)

type userRepository struct {
	client *Client
}

func NewUserRepository(client *Client) (core.UserRepository, error) {

	if err := client.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(userBucket)
		if _, err := bkt.CreateBucketIfNotExists(tgChatIdBucket); err != nil {
			return errors.Wrap(err, "cannot create tg bucket for users")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "cannot set up users buckets")
	}

	return &userRepository{
		client: client,
	}, nil
}

func (s *userRepository) SaveUser(ctx context.Context, u core.User) error {
	err := s.client.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(userBucket)
		if err := s.client.save(bkt, u.Username, u); err != nil {
			return errors.Wrapf(err, "failed to save user '%s' in bucket '%s'", u.Username, string(userBucket))
		}

		tgBkt := bkt.Bucket(tgChatIdBucket)
		tgID := strconv.FormatInt(u.TelegramID, 10)
		if err := s.client.save(tgBkt, tgID, u.Username); err != nil {
			return errors.Wrapf(err, "failed to save telegram id for user '%s' in bucket '%s'", u.Username, string(tgChatIdBucket))
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *userRepository) FindUserByUsername(ctx context.Context, username string) (core.User, error) {
	var u core.User
	err := s.client.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(userBucket)
		if err := s.client.load(bkt, username, &u); err != nil {
			return errors.Wrapf(err, "failed to load user '%s' from bucket '%s'", username, string(userBucket))
		}
		return nil
	})

	if err != nil {
		if errors.Cause(err) == errNoValue {
			return core.User{}, youpod.ErrUserNotFound
		}
		return core.User{}, err
	}

	return u, nil
}

func (s *userRepository) FindUserByTelegramID(ctx context.Context, id int64) (core.User, error) {
	var u core.User

	err := s.client.db.View(func(tx *bolt.Tx) error {

		bkt := tx.Bucket(userBucket)
		tgBkt := bkt.Bucket(tgChatIdBucket)

		tgID := strconv.FormatInt(id, 10)
		var username string
		if err := s.client.load(tgBkt, tgID, &username); err != nil {
			return errors.Wrapf(err, "failed to load username by tg id '%s' from bucket '%s'", tgID, tgChatIdBucket)
		}
		if err := s.client.load(bkt, username, &u); err != nil {
			return errors.Wrapf(err, "failed to load user '%s' from bucket '%s'", username, userBucket)
		}

		return nil
	})

	if err != nil {
		if errors.Cause(err) == errNoValue {
			return core.User{}, youpod.ErrUserNotFound
		}

		return core.User{}, err
	}

	return u, nil
}

func (s *userRepository) AddFileToUser(ctx context.Context, u core.User, fileID string) error {
	user, err := s.FindUserByUsername(ctx, u.Username)
	if err != nil {
		return errors.Wrap(err, "cannot find user")
	}
	if user.Files == nil {
		user.Files = make([]string, 0)
	}
	user.Files = append(user.Files, fileID)
	if err = s.SaveUser(ctx, user); err != nil {
		return errors.Wrap(err, "cannot update user")
	}
	return nil
}
