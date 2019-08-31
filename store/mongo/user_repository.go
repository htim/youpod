package mongo

import (
	"context"
	"github.com/htim/youpod"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	client *Client
}

func NewUserRepository(client *Client) core.UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) SaveUser(ctx context.Context, u core.User) error {
	filter := bson.D{{"username", u.Username}}
	if _, err := r.findBy(ctx, filter); err == youpod.ErrUserNotFound {
		if _, err := r.client.db.Collection(users).InsertOne(ctx, u); err != nil {
			return errors.Wrap(err, "cannot save user")
		}
		return nil
	}
	r.client.db.Collection(users).FindOneAndReplace(ctx, filter, u)
	return nil
}

func (r *userRepository) FindUserByUsername(ctx context.Context, username string) (core.User, error) {
	filter := bson.D{{"username", username}}
	user, err := r.findBy(ctx, filter)
	if err != nil {
		return core.User{}, err
	}
	return user, nil
}

func (r *userRepository) FindUserByTelegramID(ctx context.Context, id int64) (core.User, error) {
	filter := bson.D{{"telegram_id", id}}
	user, err := r.findBy(ctx, filter)
	if err != nil {
		return core.User{}, err
	}
	return user, nil
}

func (r *userRepository) AddFileToUser(ctx context.Context, u core.User, fileID string) error {
	user, err := r.FindUserByUsername(ctx, u.Username)
	if err != nil {
		return err
	}
	user.Files = append(user.Files, fileID)
	filter := bson.D{{"username", u.Username}}
	r.client.db.Collection(users).FindOneAndReplace(ctx, filter, user)
	return nil
}

func (r *userRepository) findBy(ctx context.Context, filter bson.D) (core.User, error) {
	var u core.User
	if err := r.client.db.Collection(users).FindOne(ctx, filter).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return core.User{}, youpod.ErrUserNotFound
		}
		return core.User{}, errors.Wrap(err, "cannot get user")
	}
	return u, nil
}
