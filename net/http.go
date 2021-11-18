package net

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/chindeo/pkg/net/token"
)

var (
	NetClient *Client
)

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
	Host        string            // driver redis host
	Pwd         string            // driver redis password
	Headers     map[string]string // request headers
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
		NetClient.TokenClient = &token.RedisClient{AppID: config.Appid, Host: config.Host, Pwd: config.Pwd}
	default:
		NetClient.TokenClient = &token.LocalClient{AppID: config.Appid}
	}

	NetClient.TokenClient.GetCache()

	if config.TokenDriver == "redis" && (config.Host == "") {
		return errors.New("redis driver need set redis host")
	}
	err := NetClient.TokenClient.Ping()
	if err != nil {
		return err
	}

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
	XToken string `json:"AccessToken"`
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
func (n *Client) Upload(sr *ServerResponse, name, filename string, params map[string]string, src io.Reader) ([]byte, error) {
	body := &bytes.Buffer{}                            // 初始化body参数
	writer := multipart.NewWriter(body)                // 实例化multipart
	part, err := writer.CreateFormFile(name, filename) // 创建multipart 文件字段
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, src) // 写入文件数据到multipart
	if err != nil {
		return nil, err
	}
	for key, val := range params {
		_ = writer.WriteField(key, val) // 写入body中额外参数，比如七牛上传时需要提供token
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	formcontenttype := writer.FormDataContentType()
	result := n.request(http.MethodPost, sr.FullPath, formcontenttype, body, sr.Auth)
	if len(result) == 0 {
		return result, fmt.Errorf("post %s 没有返回数据", sr.FullPath)
	}
	err = json.Unmarshal(result, sr.ResponseInfo)
	if err != nil {
		return result, fmt.Errorf("dopost: %s json.Unmarshal error：%w ,with result: %v", sr.FullPath, err, string(result))
	}

	if sr.ResponseInfo.Code == 401 {
		n.TokenClient.SetCacheToken("")
		token, err := n.GetToken()
		if err != nil {
			return result, fmt.Errorf("post %s get token err %w", sr.FullPath, err)
		}
		return result, fmt.Errorf("post %s get token %s", sr.FullPath, token)
	} else if sr.ResponseInfo.Code == 402 {
		n.TokenClient.SetCacheToken("")
		token, err := n.RfreshToken()
		if err != nil {
			return result, fmt.Errorf("post %s refresh token err %w", sr.FullPath, err)
		}
		return result, fmt.Errorf("post %s refresh token %s", sr.FullPath, token)
	} else if sr.ResponseInfo.Code != 200 {
		return result, fmt.Errorf("post [%s] 返回错误信息 [%s] 【%d】", sr.FullPath, sr.ResponseInfo.Message, sr.ResponseInfo.Code)
	}
	return result, nil
}

//POSTNet  提交数据
func (n *Client) POSTNet(sr *ServerResponse, data string) ([]byte, error) {
	result := n.request(http.MethodPost, sr.FullPath, "application/x-www-form-urlencoded; param=value", strings.NewReader(data), sr.Auth)
	if len(result) == 0 {
		return result, fmt.Errorf("post %s 没有返回数据", sr.FullPath)
	}
	err := json.Unmarshal(result, sr.ResponseInfo)
	if err != nil {
		return result, fmt.Errorf("dopost: %s json.Unmarshal error：%w ,with result: %v", sr.FullPath, err, string(result))
	}

	if sr.ResponseInfo.Code == 401 {
		n.TokenClient.SetCacheToken("")
		token, err := n.GetToken()
		if err != nil {
			return result, fmt.Errorf("post %s get token err %w", sr.FullPath, err)
		}
		return result, fmt.Errorf("post %s get token %s", sr.FullPath, token)
	} else if sr.ResponseInfo.Code == 402 {
		n.TokenClient.SetCacheToken("")
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

//GetFile  下载文件
func (n *Client) GetFile(sr *ServerResponse) ([]byte, error) {
	result := n.request("GET", sr.FullPath, "application/x-www-form-urlencoded; param=value", nil, sr.Auth)
	if len(result) == 0 {
		return result, fmt.Errorf("get %s 没有返回数据", sr.FullPath)
	}
	return result, nil
}

//GetNet  获取数据
func (n *Client) GetNet(sr *ServerResponse) ([]byte, error) {
	result := n.request("GET", sr.FullPath, "application/x-www-form-urlencoded; param=value", nil, sr.Auth)
	if len(result) == 0 {
		return result, fmt.Errorf("get %s 没有返回数据", sr.FullPath)
	}
	err := json.Unmarshal(result, sr.ResponseInfo)
	if err != nil {
		return result, fmt.Errorf("get %s 获取服务解析返回内容报错 %w", sr.FullPath, err)
	}

	if sr.ResponseInfo.Code == 401 {
		n.TokenClient.SetCacheToken("")
		_, err := n.GetToken()
		if err != nil {
			return result, fmt.Errorf("%s get token err %w", sr.FullPath, err)
		}
		return result, fmt.Errorf("get %s 返回错误信息  %s", sr.FullPath, sr.ResponseInfo.Message)
	} else if sr.ResponseInfo.Code == 402 {
		n.TokenClient.SetCacheToken("")
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
	result := n.request(http.MethodPost, n.Config.LoginUrl, "application/x-www-form-urlencoded; param=value", strings.NewReader(n.Config.LoginData), false)
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
	result := n.request("GET", n.Config.RefreshUrl, "application/x-www-form-urlencoded; param=value", nil, true)
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

func (n *Client) request(method, url, contentType string, body io.Reader, auth bool) []byte {
	result := make(chan []byte, 30)
	T := time.NewTicker(time.Duration(n.Config.TimeOver) * time.Second)
	go func() {
		t := time.Duration(n.Config.TimeOut) * time.Second
		Client := http.Client{Timeout: t}
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			result <- nil
			return
		}
		req.Header.Set("Content-Type", contentType)
		if len(n.Config.Headers) > 0 {
			for key, value := range n.Config.Headers {
				req.Header.Set(key, value)
			}
		}
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

		buf := bytes.NewBuffer(nil)
		io.Copy(buf, resp.Body)
		result <- buf.Bytes()

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
