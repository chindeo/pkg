package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type getToken struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Token `json:"data"`
}

type getServer struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    []*Server `json:"data"`
}

type getRestful struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    []*Restful `json:"data"`
}
type getDevice struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    []*Device `json:"data"`
}

type Req struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Token struct {
	XToken string `json:"X-Token"`
}

type Server struct {
	Id            int64  `json:"id"`
	Ip            string `json:"ip"`
	Port          int64  `json:"port"`
	Account       string `json:"account"`
	Pwd           string `json:"pwd"`
	ServiceTypeId int64  `json:"service_type_id"`
	ServiceName   string `json:"service_name"`
	ServiceTitle  string `json:"service_title"`
}

type Restful struct {
	Id  int64  `json:"id"`
	Url string `json:"url"`
}

type Time struct {
	time.Time
}

// returns time.Now() no matter what!
func (t *Time) UnmarshalJSON(b []byte) error {
	// you can now parse b as thoroughly as you want

	*t = Time{time.Now()}
	return nil
}

type Device struct {
	IsError   int64  `json:"is_error" `
	DevStatus int64  `json:"device_status"`
	DevCode   string `json:"device_code"`
	LogAt     Time   `json:"log_at"`
}

//http://fyxt.t.chindeo.com/platform/report/getService  获取服务
func GetServices(appid, appsecret, host string, timeover, timeout int64) ([]*Server, error) {
	re := &getServer{}
	fullUrl := host + "platform/report/getService"
	result := Request(appid, "GET", fullUrl, "", timeover, timeout, true)
	if len(result) == 0 {
		return nil, errors.New(fmt.Sprintf("GetServices 获取服务请求没有返回数据"))
	}
	err := json.Unmarshal(result, re)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetServices 获取服务解析返回内容报错 :%v", err))
	}

	if re.Code == 200 {
		return re.Data, nil
	} else if re.Code == 401 {
		err = GetToken(appid, appsecret, host, timeover, timeout)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("get token err %v", err))
		}
		return nil, errors.New("重新获取 token")
	} else if re.Code == 402 {
		err = RfreshToken(appid, host, timeover, timeout)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("refresh token err %v", err))
		}
		return nil, errors.New("刷新 token")
	} else {
		return nil, errors.New(fmt.Sprintf("GetServices 获取服务返回错误信息 :%v", re.Message))
	}
}

//http://fyxt.t.chindeo.com/platform/report/getRestful  获取接口列表
func GetRestfuls(appid, appsecret, host string, timeover, timeout int64) ([]*Restful, error) {
	var re getRestful
	fullUrl := host + "platform/report/getRestful"
	result := Request(appid, "GET", fullUrl, "", timeover, timeout, true)
	if len(result) == 0 {
		return nil, errors.New(fmt.Sprintf("GetRestfuls 获取接口请求没有返回数据"))
	}
	err := json.Unmarshal(result, &re)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetRestfuls 获取接口解析返回内容报错 :%v", err))
	}

	if re.Code == 200 {
		return re.Data, nil
	} else if re.Code == 401 {
		err = GetToken(appid, appsecret, host, timeover, timeout)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("get token err %v", err))
		}
		return nil, errors.New("重新获取 token")
	} else if re.Code == 402 {
		err = RfreshToken(appid, host, timeover, timeout)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("refresh token err %v", err))
		}
		return nil, errors.New("刷新 token")
	} else {
		return nil, errors.New(fmt.Sprintf("GetRestfuls 获取接口返回错误信息 :%v", re.Message))
	}
}

//http://fyxt.t.chindeo.com/platform/report/getDevice  获取设备列表
func GetDevices(appid, appsecret, host string, timeover, timeout int64) ([]*Device, error) {
	var re getDevice
	fullUrl := host + "platform/report/getDevice"
	result := Request(appid, "GET", fullUrl, "", timeover, timeout, true)
	if len(result) == 0 {
		return nil, errors.New(fmt.Sprintf("GetRestfuls 获取设备请求没有返回数据"))
	}
	err := json.Unmarshal(result, &re)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetRestfuls 获取设备解析返回内容报错 :%v", err))
	}

	if re.Code == 200 {
		return re.Data, nil
	} else if re.Code == 401 {
		err = GetToken(appid, appsecret, host, timeover, timeout)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("get token err %v", err))
		}
		return nil, errors.New("重新获取 token")
	} else if re.Code == 402 {
		err = RfreshToken(appid, host, timeover, timeout)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("refresh token err %v", err))
		}
		return nil, errors.New("刷新 token")
	} else {
		return nil, errors.New(fmt.Sprintf("GetRestfuls 获取设备返回错误信息 :%v", re.Message))
	}
}

//http://fyxt.t.chindeo.com/platform/report/device  发送设备日志信息
//http://fyxt.t.chindeo.com/platform/report/service  提交服务监控信息
func SyncServices(appid, appsecret, host, path, data string, timeover, timeout int64) (interface{}, error) {
	var re Req
	result := Request(appid, "POST", host+path, data, timeover, timeout, true)
	if len(result) == 0 {
		return nil, errors.New(fmt.Sprintf("SyncServices 同步数据请求没有返回数据"))
	}
	err := json.Unmarshal(result, &re)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("SyncServices dopost: %s json.Unmarshal error：%v ,with result: %v", path, err, string(result)))
	}

	if re.Code == 200 {
		return re, nil
	} else if re.Code == 401 {
		err = GetToken(appid, appsecret, host, timeover, timeout)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("get token err %v", err))
		}
		return nil, errors.New("重新获取 token")
	} else if re.Code == 402 {
		err = RfreshToken(appid, host, timeover, timeout)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("refresh token err %v", err))
		}
		return nil, errors.New("刷新 token")
	} else {
		return nil, errors.New(fmt.Sprintf("SyncServices 获取接口返回错误信息 :%v", re.Message))
	}
}

//http://fyxt.t.chindeo.com/platform/application/login
//http://fyxt.t.chindeo.com/platform/report/device
func GetToken(appid, appsecret, host string, timeover, timeout int64) error {
	var re getToken
	fullUrl := host + "platform/application/login"
	data := fmt.Sprintf("appid=%s&appsecret=%s&apptype=%s", appid, appsecret, "hospital")
	result := Request(appid, "POST", fullUrl, data, timeover, timeout, false)
	if len(result) == 0 {
		return errors.New("请求没有返回数据")
	}

	err := json.Unmarshal(result, &re)
	if err != nil {
		return err
	}

	if re.Code == 200 {
		SetCacheToken(re.Data.XToken)
		return nil
	} else {
		return errors.New(re.Message)
	}
}

//http://fyxt.t.chindeo.com/platform/application/update_token
//http://fyxt.t.chindeo.com/platform/report/device
func RfreshToken(appid, host string, timeover, timeout int64) error {
	var re getToken
	fullUrl := host + "platform/application/update_token"
	result := Request(appid, "GET", fullUrl, "", timeover, timeout, true)
	if len(result) == 0 {
		return errors.New("请求没有返回数据")
	}
	err := json.Unmarshal(result, &re)
	if err != nil {
		return err
	}
	if re.Code == 200 {
		SetCacheToken(re.Data.XToken)
		return nil
	} else {
		return errors.New(re.Message)
	}
}

func Request(appid, method, fullUrl, data string, timeover, timeout int64, auth bool) []byte {
	var result = make(chan []byte, 10)
	T := time.Tick(time.Duration(timeover) * time.Second)
	go func() {
		t := time.Duration(timeout) * time.Second
		Client := http.Client{Timeout: t}
		req, err := http.NewRequest(method, fullUrl, strings.NewReader(data))
		if err != nil {
			fmt.Println(fmt.Sprintf("%s: %+v", fullUrl, err))
			result <- nil
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		if auth {
			req.Header.Set("X-Token", GetCacheToken())
			phpSessionId := GetSessionId(appid)
			if phpSessionId != nil {
				req.AddCookie(phpSessionId)
			}
		}
		var resp *http.Response
		resp, err = Client.Do(req)
		if err != nil {
			fmt.Println(fmt.Sprintf("%s: %+v", fullUrl, err))
			result <- nil
			return
		}
		defer resp.Body.Close()

		if !auth {
			SetSessionId(resp.Cookies(), appid)
		}

		b, _ := ioutil.ReadAll(resp.Body)
		result <- b

	}()

	for {
		select {
		case x := <-result:
			return x
		case <-T:
			return nil
		}
	}

}
