package rds

import (
	"github.com/go-redis/redis/v8"
	"time"
)

type Config struct {
	Addr        string
	Password    string
	Bb          int
	Idle        int
	Active      int
	IdleTimeout time.Duration
}

func New(c *Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password, // no password set
		DB:       c.Bb,       // use default DB
	})
}
