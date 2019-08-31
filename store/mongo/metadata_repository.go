package mongo

import (
	"context"
	"github.com/htim/youpod"
	"github.com/htim/youpod/core"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type metadataRepository struct {
	client *Client
}

func NewMetadataRepository(client *Client) core.MetadataRepository {
	return &metadataRepository{client: client}
}

func (r *metadataRepository) GetFileMetadata(ctx context.Context, ID string) (core.Metadata, error) {
	var m core.Metadata

	filter := bson.D{{"file_id", ID}}

	if err := r.client.db.Collection(metadata).FindOne(
		ctx,
		filter,
	).Decode(&m); err != nil {

		if err == mongo.ErrNoDocuments {
			return core.Metadata{}, youpod.ErrMetadataNotFound
		}

		return core.Metadata{}, errors.Wrap(err, "cannot find metadata")
	}

	return m, nil
}

func (r *metadataRepository) SaveFileMetadata(ctx context.Context, m core.Metadata) (err error) {
	if _, err := r.client.db.Collection(metadata).InsertOne(ctx, m); err != nil {
		return errors.Wrap(err, "cannot save metadata")
	}
	return nil
}
