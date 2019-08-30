package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	dbname   = "youpod_db"
	collname = "youpod"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.client.Connect(ctx); err != nil {
		return err
	}
	return nil
}
