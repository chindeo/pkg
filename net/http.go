package net

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var NetClient *Client

type Client struct {
	Config *Config
}

type Config struct {
	Appid      string
	AppSecret  string
	LoginData  string
	LoginUrl   string
	RefreshUrl string
	TimeOver   int64
	TimeOut    int64
}

func NewNetClient(config *Config) {
	if NetClient != nil {
		return
	}
	NetClient = &Client{config}
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
	result := Request(n.Config.Appid, "POST", sr.FullPath, data, n.Config.TimeOver, n.Config.TimeOut, sr.Auth)
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
	result := Request(n.Config.Appid, "GET", sr.FullPath, "", n.Config.TimeOver, n.Config.TimeOut, sr.Auth)
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
	re := &responseToken{}
	result := Request(n.Config.Appid, "POST", n.Config.LoginUrl, n.Config.LoginData, n.Config.TimeOver, n.Config.TimeOut, false)
	if len(result) == 0 {
		return "", fmt.Errorf("GetToken  %s get empty data", n.Config.LoginUrl)
	}

	err := json.Unmarshal(result, re)
	if err != nil {
		return "", fmt.Errorf("unmarshal json %s error %w", string(result), err)
	}

	if re.Code == 200 {
		SetCacheToken(re.Data.XToken, n.Config.Appid)
		return re.Data.XToken, nil
	} else {
		return "", fmt.Errorf(re.Message)
	}
}

// RfreshToken
func (n *Client) RfreshToken() (string, error) {
	re := &responseToken{}
	result := Request(n.Config.Appid, "GET", n.Config.RefreshUrl, "", n.Config.TimeOver, n.Config.TimeOut, true)
	if len(result) == 0 {
		return "", fmt.Errorf("RfreshToken  %s get empty data", n.Config.RefreshUrl)
	}
	err := json.Unmarshal(result, &re)
	if err != nil {
		return "", fmt.Errorf("unmarshal json %s error %w", string(result), err)
	}
	if re.Code == 200 {
		SetCacheToken(re.Data.XToken, n.Config.Appid)
		return re.Data.XToken, nil
	} else {
		return "", errors.New(re.Message)
	}
}

func Request(appid, method, url, data string, timeover, timeout int64, auth bool) []byte {
	result := make(chan []byte, 30)
	T := time.NewTicker(time.Duration(timeover) * time.Second)
	go func() {
		t := time.Duration(timeout) * time.Second
		Client := http.Client{Timeout: t}
		req, err := http.NewRequest(method, url, strings.NewReader(data))
		if err != nil {
			result <- nil
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		if auth && appid != "" {
			req.Header.Set("X-Token", GetCacheToken(appid))
			phpSessionId := GetSessionId(appid)
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

		if !auth && appid != "" {
			SetSessionId(resp.Cookies(), appid)
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
