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

type getToken struct {
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
