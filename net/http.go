package net

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/chindeo/pkg/net/token"
)

var NetClient *Client

type Client struct {
	Config      *Config
	TokenClient token.TokenClient
}

type Config struct {
	Appid       string
	AppSecret   string
	LoginData   string
	LoginUrl    string
	RefreshUrl  string
	TimeOver    int64
	TimeOut     int64
	TokenDriver string
	Host        string
	Pwd         string
}

func NewNetClient(config *Config) error {
	if NetClient != nil {
		return nil
	}
	NetClient = &Client{Config: config}
	switch config.TokenDriver {
	case "local":
		NetClient.TokenClient = &token.LocalClient{AppID: config.Appid}
	case "redis":
		if config.Host == "" || config.Pwd == "" {
			return errors.New("redis driver need set redis host and password")
		}
		NetClient.TokenClient = &token.RedisClient{AppID: config.Appid, Host: config.Host, Pwd: config.Pwd}
	default:
		NetClient.TokenClient = &token.LocalClient{AppID: config.Appid}
	}
	NetClient.TokenClient.GetCache()
	return nil
}

type responseToken struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Token `json:"data"`
}

type Req struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Token struct {
	XToken string `json:"X-Token"`
}

type ServerResponse struct {
	FullPath     string
	Auth         bool
	ResponseInfo *ResponseInfo
}

type ResponseInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//POSTNet  提交数据
func (n *Client) POSTNet(sr *ServerResponse, data string) ([]byte, error) {
	result := n.request("POST", sr.FullPath, data, sr.Auth)
	if len(result) == 0 {
		return result, fmt.Errorf("post %s 没有返回数据", sr.FullPath)
	}
	err := json.Unmarshal(result, sr.ResponseInfo)
	if err != nil {
		return result, fmt.Errorf("dopost: %s json.Unmarshal error：%w ,with result: %v", sr.FullPath, err, string(result))
	}

	if sr.ResponseInfo.Code == 401 {
		token, err := n.GetToken()
		if err != nil {
			return result, fmt.Errorf("post %s get token err %w", sr.FullPath, err)
		}
		return result, fmt.Errorf("post %s get token %s", sr.FullPath, token)
	} else if sr.ResponseInfo.Code == 402 {
		token, err := n.RfreshToken()
		if err != nil {
			return result, fmt.Errorf("post %s refresh token err %w", sr.FullPath, err)
		}
		return result, fmt.Errorf("post %s refresh token %s", sr.FullPath, token)
	} else if sr.ResponseInfo.Code != 200 {
		return result, fmt.Errorf("post %s 返回错误信息 %s 【%d】", sr.FullPath, sr.ResponseInfo.Message, sr.ResponseInfo.Code)
	}
	return result, nil
}

//GetNet  获取数据
func (n *Client) GetNet(sr *ServerResponse) ([]byte, error) {
	result := n.request("GET", sr.FullPath, "", sr.Auth)
	if len(result) == 0 {
		return result, fmt.Errorf("get %s 没有返回数据", sr.FullPath)
	}
	err := json.Unmarshal(result, sr.ResponseInfo)
	if err != nil {
		return result, fmt.Errorf("get %s 获取服务解析返回内容报错 %w", sr.FullPath, err)
	}

	if sr.ResponseInfo.Code == 401 {
		token, err := n.GetToken()
		if err != nil {
			return result, fmt.Errorf("%s get token err %w", sr.FullPath, err)
		}
		return result, fmt.Errorf("get %s 返回错误信息 get token %s", sr.FullPath, token)
	} else if sr.ResponseInfo.Code == 402 {
		token, err := n.RfreshToken()
		if err != nil {
			return result, fmt.Errorf("get %s refresh token err %w", sr.FullPath, err)
		}
		return result, fmt.Errorf("get %s 返回错误信息 refresh token %s ", sr.FullPath, token)
	} else if sr.ResponseInfo.Code != 200 {
		return result, fmt.Errorf("get %s 返回错误信息 %s 【%d】", sr.FullPath, sr.ResponseInfo.Message, sr.ResponseInfo.Code)
	}
	return result, nil
}

// GetToken
// data := fmt.Sprintf("appid=%s&appsecret=%s&apptype=%s", appid, appsecret, "hospital")
func (n *Client) GetToken() (string, error) {
	token := n.TokenClient.GetCacheToken()
	if token != "" {
		return token, nil
	}
	re := &responseToken{}
	result := n.request("POST", n.Config.LoginUrl, n.Config.LoginData, false)
	if len(result) == 0 {
		return "", fmt.Errorf("GetToken  %s get empty data", n.Config.LoginUrl)
	}

	err := json.Unmarshal(result, re)
	if err != nil {
		return "", fmt.Errorf("unmarshal json %s error %w", string(result), err)
	}

	if re.Code == 200 {
		n.TokenClient.SetCacheToken(re.Data.XToken)
		return re.Data.XToken, nil
	} else {
		return "", fmt.Errorf(re.Message)
	}
}

// RfreshToken
func (n *Client) RfreshToken() (string, error) {
	re := &responseToken{}
	result := n.request("GET", n.Config.RefreshUrl, "", true)
	if len(result) == 0 {
		return "", fmt.Errorf("RfreshToken  %s get empty data", n.Config.RefreshUrl)
	}
	err := json.Unmarshal(result, &re)
	if err != nil {
		return "", fmt.Errorf("unmarshal json %s error %w", string(result), err)
	}
	if re.Code == 200 {
		n.TokenClient.SetCacheToken(re.Data.XToken)
		return re.Data.XToken, nil
	} else {
		return "", errors.New(re.Message)
	}
}

func (n *Client) request(method, url, data string, auth bool) []byte {
	result := make(chan []byte, 30)
	T := time.NewTicker(time.Duration(n.Config.TimeOver) * time.Second)
	go func() {
		t := time.Duration(n.Config.TimeOut) * time.Second
		Client := http.Client{Timeout: t}
		req, err := http.NewRequest(method, url, strings.NewReader(data))
		if err != nil {
			result <- nil
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		if auth && n.Config.Appid != "" {
			req.Header.Set("X-Token", n.TokenClient.GetCacheToken())
			phpSessionId := n.TokenClient.GetSessionId()
			if phpSessionId != nil {
				req.AddCookie(phpSessionId)
			}
		}
		var resp *http.Response
		resp, err = Client.Do(req)
		if err != nil {
			result <- nil
			return
		}
		defer resp.Body.Close()

		if !auth && n.Config.Appid != "" {
			n.TokenClient.SetSessionId(resp.Cookies())
		}

		b, _ := ioutil.ReadAll(resp.Body)
		result <- b

	}()

	for {
		select {
		case x := <-result:
			return x
		case <-T.C:
			return []byte("请求超时")
		}
	}
}
