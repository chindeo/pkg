package redis_client

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
)

func Test_NewClient(t *testing.T) {
	t.Run("new redis client", func(t *testing.T) {
		redis := NewClient(&redis.Options{Addr: "127.0.0.1:6379", Password: "snowlyg", DB: 0})
		err := redis.Ping(context.Background()).Err()
		if err != nil {
			t.Errorf("redis ping is fault,get msg %s", err.Error())
		}
	})
}
