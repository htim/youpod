package bolt

import (
	"encoding/json"
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

// save marshaled value to key for bucket. Should run in update tx
func (c *Client) save(bkt *bolt.Bucket, key string, value interface{}) (err error) {
	if value == nil {
		return errors.Errorf("cannot save nil value for %s", key)
	}
	jdata, jerr := json.Marshal(value)
	if jerr != nil {
		return errors.Wrap(jerr, "cannot marshal value")
	}
	if err = bkt.Put([]byte(key), jdata); err != nil {
		return errors.Wrapf(err, "failed to save key %s", key)
	}
	return nil
}

// load and unmarshal json value by key from bucket. Should run in view tx
func (c *Client) load(bkt *bolt.Bucket, key string, res interface{}) error {
	value := bkt.Get([]byte(key))
	if value == nil {
		return errNoValue
	}

	if err := json.Unmarshal(value, &res); err != nil {
		return errors.Wrap(err, "failed to unmarshal")
	}
	return nil
}
