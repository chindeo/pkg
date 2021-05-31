package token

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/chindeo/pkg/net/redis_client"
	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

var (
	onceLc sync.Once
)

type RedisClient struct {
	AppID   string
	Host    string
	Pwd     string
	rw      sync.RWMutex
	token   string
	ca      *redis.Client
	phpsess *http.Cookie
}

func (lc *RedisClient) SetSessionId(cookies []*http.Cookie) {
	for _, cookie := range cookies {
		if cookie.Name == "PHPSESSID" {
			lc.ca.Set(context.Background(), fmt.Sprintf("PHPSESSIONID_%s", lc.AppID), cookie, cache.DefaultExpiration)
		}
	}
}

func (lc *RedisClient) GetSessionId() *http.Cookie {
	if lc.phpsess != nil {
		return lc.phpsess
	}
	var hc *http.Cookie
	err := lc.ca.Get(context.Background(), fmt.Sprintf("PHPSESSIONID_%s", lc.AppID)).Scan(hc)
	if err != nil {
		lc.phpsess = hc
	}
	return lc.phpsess
}

func (lc *RedisClient) GetCache() {
	onceLc.Do(func() {
		lc.ca = redis_client.NewClient(&redis.Options{
			Addr:     lc.Host,
			Password: lc.Pwd, // no password set
			DB:       0,      // use default DB
		})

	})
}

func (lc *RedisClient) SetCacheToken(token string) {
	lc.rw.Lock()
	defer lc.rw.Unlock()
	lc.ca.Set(context.Background(), "XToken:"+lc.AppID, token, cache.DefaultExpiration)
}

func (lc *RedisClient) GetCacheToken() string {
	lc.rw.RLock()
	defer lc.rw.RUnlock()
	foo, err := lc.ca.Get(context.Background(), "XToken:"+lc.AppID).Result()
	if err != nil {
		lc.token = ""
	}
	lc.token = foo
	return lc.token
}

func (lc *RedisClient) Ping() error {
	return lc.ca.Ping(context.Background()).Err()
}
