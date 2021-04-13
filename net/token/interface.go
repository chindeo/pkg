package token

import (
	"net/http"
)

type TokenClient interface {
	SetSessionId(cookies []*http.Cookie)
	GetSessionId() *http.Cookie
	GetCache()
	SetCacheToken(token string)
	GetCacheToken() string
}
