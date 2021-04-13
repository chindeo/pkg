package token

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

type LocalClient struct {
	AppID   string
	rw      sync.RWMutex
	token   string
	ca      *cache.Cache
	phpsess *http.Cookie
}

func (lc *LocalClient) SetSessionId(cookies []*http.Cookie) {
	for _, cookie := range cookies {
		if cookie.Name == "PHPSESSID" {
			lc.ca.Set(fmt.Sprintf("PHPSESSIONID_%s", lc.AppID), cookie, cache.DefaultExpiration)
		}
	}
}

func (lc *LocalClient) GetSessionId() *http.Cookie {
	if lc.phpsess != nil {
		return lc.phpsess
	}
	foo, found := lc.ca.Get(fmt.Sprintf("PHPSESSIONID_%s", lc.AppID))
	if found {
		lc.phpsess = foo.(*http.Cookie)
	}
	return lc.phpsess
}

func (lc *LocalClient) GetCache() {
	if lc.ca != nil {
		return
	}
	lc.ca = cache.New(24*time.Hour, 7*24*time.Hour)
}

func (lc *LocalClient) SetCacheToken(token string) {
	lc.rw.Lock()
	defer lc.rw.Unlock()
	lc.ca.Set("XToken:"+lc.AppID, token, cache.DefaultExpiration)
}

func (lc *LocalClient) GetCacheToken() string {
	lc.rw.RLock()
	defer lc.rw.RUnlock()
	foo, found := lc.ca.Get("XToken:" + lc.AppID)
	if found {
		lc.token = foo.(string)
	}
	return lc.token
}
