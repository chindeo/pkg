package redis_client

import (
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	once        *sync.Once
)

func NewClient(options *redis.Options) *redis.Client {
	once.Do(func() {
		redisClient = redis.NewClient(options)
	})
	return redisClient
}
