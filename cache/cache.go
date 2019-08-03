package cache

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
)

type LoadingCache struct {
	lru *lru.Cache
}

func NewLoadingCache() (*LoadingCache, error) {
	c, err := lru.New(10)
	if err != nil {
		return nil, err
	}
	return &LoadingCache{lru: c}, nil
}

func (c *LoadingCache) GetOrLoad(key interface{}, loadFunc func(key interface{}) (interface{}, error)) (interface{}, error) {
	if v, ok := c.lru.Get(key); ok {
		return v, nil
	}

	v, err := loadFunc(key)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load value for key '%v'", key)
	}
	c.lru.Add(key, v)
	return v, nil
}

func (c *LoadingCache) Get(key interface{}) (interface{}, bool) {
	return c.lru.Get(key)
}

func (c *LoadingCache) Add(key interface{}, value interface{}) {
	c.lru.Add(key, value)
}
