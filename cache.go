package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/gohp/cache/rds"
	lru "github.com/hashicorp/golang-lru"
	"sync"
)

type Storage interface {
	Get(key string) (string, error)
	Set(key string, value string) error
	Delete(key string) error
	Close() error
}

type StWithCache struct {
	store Storage
	cache *lru.Cache
	lock  sync.RWMutex
}

func InitCache(c *rds.Config, cacheSize int) *StWithCache {
	cache, _ := lru.New(cacheSize)
	storage, _ := NewRdsStorage(c)
	r := &StWithCache{
		cache: cache,
		store: storage,
	}
	r.store = storage
	return r
}

func (s *StWithCache) Get(key string) (string, error) {
	// local cache get
	c, exist := s.cache.Get(key)
	if exist {
		return c.(string), nil
	}
	// redis cache get
	s.lock.Lock()
	v, err := s.store.Get(key)
	s.lock.Unlock()
	if v != "" {
		// grab data from redis and write into local cache
		s.cache.Add(key, v)
	}
	return v, err
}

func (s *StWithCache) Set(key string, value string) error {
	old, exist := s.cache.Get(key)
	s.cache.Add(key, value)
	s.lock.Lock()
	err := s.store.Set(key, value)
	s.lock.Unlock()
	if err != nil && exist {
		s.cache.Add(key, old)
	}
	return err
}

func (s *StWithCache) Delete(key string) error {
	old, exist := s.cache.Get(key)

	s.cache.Remove(key)
	s.lock.Lock()
	err := s.store.Delete(key)
	s.lock.Unlock()

	if err != nil && exist {
		s.cache.Add(key, old)
	}
	return err
}

func (s *StWithCache) Close() error {
	return s.store.Close()
}

type RdsStorage struct {
	pool *redis.Client
}

func NewRdsStorage(c *rds.Config) (Storage, error) {
	pool := rds.New(c)
	r := &RdsStorage{
		pool: pool,
	}

	return r, nil
}

func (s *RdsStorage) Get(key string) (string, error) {
	val, err := s.pool.Get(context.TODO(), key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (s *RdsStorage) Set(key string, value string) error {
	_, err := s.pool.Set(context.TODO(), key, value, 0).Result()
	// TODO set error return old value
	return err
}

func (s *RdsStorage) Delete(key string) error {
	_, err := s.pool.Del(context.TODO(), key).Result()
	return err
}

func (s *RdsStorage) Close() error {
	return s.pool.Close()
}
