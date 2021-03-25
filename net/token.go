package net

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

var ca *cache.Cache
var token string
var phpsess *http.Cookie
var rw sync.RWMutex

func SetSessionId(cookies []*http.Cookie, appid string) {
	for _, cookie := range cookies {
		if cookie.Name == "PHPSESSID" {
			GetCache().Set(fmt.Sprintf("PHPSESSIONID_%s", appid), cookie, cache.DefaultExpiration)
		}
	}
}

func GetSessionId(appid string) *http.Cookie {
	if phpsess != nil {
		return phpsess
	}
	foo, found := GetCache().Get(fmt.Sprintf("PHPSESSIONID_%s", appid))
	if found {
		phpsess = foo.(*http.Cookie)
	}
	return phpsess
}

func GetCache() *cache.Cache {
	if ca != nil {
		return ca
	}
	ca = cache.New(1*time.Hour, 2*time.Hour)
	return ca
}

func SetCacheToken(token, appid string) {
	rw.Lock()
	defer rw.Unlock()
	GetCache().Set("XToken:"+appid, token, cache.DefaultExpiration)
}

func GetCacheToken(appid string) string {
	rw.RLock()
	defer rw.RUnlock()
	foo, found := GetCache().Get("XToken:" + appid)
	if found {
		token = foo.(string)
	}
	return token
}
