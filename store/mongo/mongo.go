package mongo

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	users    = "users"
	metadata = "metadata"
)

type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewClient(uri string) (*Client, error) {
	mClient, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: mClient,
		db:     mClient.Database("youpod"),
	}, nil
}

func (c *Client) Open() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.client.Connect(ctx); err != nil {
		return err
	}

	usersIndexes := []mongo.IndexModel{
		{
			Keys: bson.M{
				"telegram_id": 1,
			},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{
				"username": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := c.db.Collection(users).Indexes().CreateMany(ctx, usersIndexes); err != nil {
		return errors.Wrap(err, "cannot create indexes on users collection")
	}

	if _, err := c.db.Collection(metadata).Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"file_id": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	); err != nil {
		return errors.Wrap(err, "cannot create index on metadata collection")
	}

	return nil
}
