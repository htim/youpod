package bolt

import (
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	userBucket  = []byte("users")
	filesBucket = []byte("files")
	folders     = []byte("folders")
)

var (
	errNoValue = errors.New("no value")
)

var (
	txKey = "bolt:tx_ctx_key"
)

type Client struct {
	path string
	db   *bolt.DB
}

func NewClient(path string) *Client {
	return &Client{
		path: path,
	}
}

func (c *Client) Open() error {
	db, err := bolt.Open(c.path, 0666, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return errors.Wrapf(err, "cannot open db: %s", c.path)
	}
	c.db = db

	tx, err := c.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	topBuckets := [][]byte{
		userBucket,
		filesBucket,
		folders,
	}

	for _, b := range topBuckets {
		if _, err := tx.CreateBucketIfNotExists(b); err != nil {
			return errors.Wrapf(err, "cannot create bucket: %s", string(b))
		}
	}

	return tx.Commit()
}

func (c *Client) Close() error {
	return c.db.Close()
}
