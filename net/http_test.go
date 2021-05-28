package net

import (
	"fmt"
	"testing"
)

var (
	Host       string = "http://op.t.chindeo.com"
	Appid      string = "XvtbOYC0VQEweLqW01cyHvn8TTDq9Yun"
	Appsecret  string = "iemkUnyvJLghZgkKJBDaB4r4hNlFkab673BO9w6AuU367Rqv5N"
	Loginuri   string = "/platform/application/login"
	Refreshuri string = "/platform/application/update_token"
	Timeover   int64  = 5
	Timeout    int64  = 10
	Restfuluri string = "/platform/report/restful"
	Serviceuri string = "/platform/report/service"
	Deviceuri  string = "/platform/report/device"
)

func Test_GetNet(t *testing.T) {
	loginData := fmt.Sprintf("appid=%s&appsecret=%s&apptype=%s", Appid, Appsecret, "hospital")
	NewNetClient(&Config{
		Appid:      Appid,
		AppSecret:  Appsecret,
		LoginUrl:   Host + Loginuri,
		RefreshUrl: Host + Refreshuri,
		LoginData:  loginData,
		TimeOver:   Timeover,
		TimeOut:    Timeout,
	})

	serviceResponseRestful := &ServerResponse{
		FullPath:     Host + "/platform/report/getRestful",
		Auth:         true,
		ResponseInfo: &ResponseInfo{},
	}

	serviceResponseService := &ServerResponse{
		FullPath:     Host + "/platform/report/getService",
		Auth:         true,
		ResponseInfo: &ResponseInfo{},
	}
	serviceResponseDevice := &ServerResponse{
		FullPath:     Host + "/platform/report/getDevice",
		Auth:         true,
		ResponseInfo: &ResponseInfo{},
	}

	tests := []struct {
		name         string
		responseInfo *ServerResponse
	}{
		{
			name:         "获取接口列表",
			responseInfo: serviceResponseRestful,
		},
		{
			name:         "获取服务列表",
			responseInfo: serviceResponseService,
		},
		{
			name:         "获取设备列表",
			responseInfo: serviceResponseDevice,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NetClient.GetNet(tt.responseInfo)
			if err != nil {
				t.Errorf("GetNet() error = %v", err)
				return
			}
			if b == nil {
				t.Errorf("GetNet() got nil")
			}
		})
	}
}

func Test_POSTNet(t *testing.T) {
	loginData := fmt.Sprintf("appid=%s&appsecret=%s&apptype=%s", Appid, Appsecret, "hospital")
	NewNetClient(&Config{
		Appid:      Appid,
		AppSecret:  Appsecret,
		LoginUrl:   Host + Loginuri,
		RefreshUrl: Host + Refreshuri,
		LoginData:  loginData,
		TimeOver:   Timeover,
		TimeOut:    Timeout,
	})
	serviceResponseRestful := &ServerResponse{
		FullPath:     Host + Restfuluri,
		Auth:         true,
		ResponseInfo: &ResponseInfo{},
	}

	serviceResponseService := &ServerResponse{
		FullPath:     Host + Serviceuri,
		Auth:         true,
		ResponseInfo: &ResponseInfo{},
	}
	serviceResponseDevice := &ServerResponse{
		FullPath:     Host + Deviceuri,
		Auth:         true,
		ResponseInfo: &ResponseInfo{},
	}

	dataRestful := fmt.Sprintf("restful_data=%s", string(""))
	dataService := fmt.Sprintf("fault_data=%s", string(""))
	dataDevice := fmt.Sprintf("log_msgs=%s", string(""))
	tests := []struct {
		name         string
		responseInfo *ServerResponse
		data         string
	}{
		{
			name:         "同步接口列表故障数据",
			responseInfo: serviceResponseRestful,
			data:         dataRestful,
		},
		{
			name:         "同步服务列表故障数据",
			responseInfo: serviceResponseService,
			data:         dataService,
		},
		{
			name:         "同步设备故障数据",
			responseInfo: serviceResponseDevice,
			data:         dataDevice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NetClient.POSTNet(tt.responseInfo, tt.data)
			if err != nil {
				t.Errorf("POSTNet() error = %v", err)
				return
			}
		})
	}
}

func Test_NewNetClient(t *testing.T) {
	args := []*Config{
		{
			Appid:       "Appid",
			AppSecret:   "AppSecret",
			LoginData:   "LoginData",
			LoginUrl:    "LoginUrl",
			RefreshUrl:  "RefreshUrl",
			TimeOver:    5,
			TimeOut:     5,
			TokenDriver: "redis",
			Host:        "127.0.0.1:6379", // driver redis host
			Pwd:         "snowlyg",        // driver redis password
			Headers:     nil,              // request headers
		},
		{
			Appid:       "Appid",
			AppSecret:   "AppSecret",
			LoginData:   "LoginData",
			LoginUrl:    "LoginUrl",
			RefreshUrl:  "RefreshUrl",
			TimeOver:    5,
			TimeOut:     5,
			TokenDriver: "lcoal",
			Host:        "",  // driver redis host
			Pwd:         "",  // driver redis password
			Headers:     nil, // request headers
		},
	}
	for _, arg := range args {
		t.Run("new redis client", func(t *testing.T) {
			err := NewNetClient(arg)
			if err != nil {
				t.Errorf("redis ping is fault,get msg %s", err.Error())
			}
		})
	}

}
