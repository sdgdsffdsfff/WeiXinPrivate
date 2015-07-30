package wapi

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/bitly/go-simplejson"
)

const (
	appid            = "wxafc2e7724ecc19e9"
	appsecret        = "e0c95308a69c57355064aaf8de85c59d"
	access_token_url = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET"
	server_ips_url   = "https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=ACCESS_TOKEN"
)

// {"access_token":"ACCESS_TOKEN","expires_in":7200}
type AccessToken struct {
	Access_token string
	Expires_in   int
}

// {"errcode":40013,"errmsg":"invalid appid"}
func transErrcode(errcode, errmsg string) error {
	// TODO 转义全局错误代码
	return fmt.Errorf("weixin errcode:%s errmsg:%s", errcode, errmsg)
}

func GetAcessToken() (*AccessToken, error) {
	url := access_token_url
	url = strings.Replace(url, "APPID", appid, 1)
	url = strings.Replace(url, "APPSECRET", appsecret, 1)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get acess_token err\nhttp status:%d %s\nerror:%s\n", resp.StatusCode, resp.Status, err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp error:%s\n", err.Error())
	}

	log.Printf("weixin access_token return %s\n", string(body))

	json, err := simplejson.NewJson(body)
	if err != nil {
		return nil, fmt.Errorf("format to json err:%s\n", err.Error())
	}
	access_token, err := json.Get("access_token").String()
	if err == nil {
		// 正常返回
		expires_in, err := json.Get("expires_in").Int()
		if err != nil {
			return nil, fmt.Errorf("not found expires_in.")
		}
		return &AccessToken{access_token, expires_in}, nil
	} else {
		// 错误返回
		errcode, err := json.Get("errcode").String()
		if err != nil {
			return nil, fmt.Errorf("not found errcode.")
		}
		errmsg, err := json.Get("errmsg").String()
		if err != nil {
			return nil, fmt.Errorf("not found errmsg.")
		}
		return nil, transErrcode(errcode, errmsg)
	}
}

func GetWeixinServerIP(access_token string) ([]string, error) {
	url := server_ips_url
	url = strings.Replace(url, "ACCESS_TOKEN", access_token, 1)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get ip_list err\nhttp status:%d %s\nerror:%s\n", resp.StatusCode, resp.Status, err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp error:%s\n", err.Error())
	}

	log.Printf("weixin ServerIP return %s\n", string(body))

	json, err := simplejson.NewJson(body)
	if err != nil {
		return nil, fmt.Errorf("format to json err:%s\n", err.Error())
	}
	ips, err := json.Get("ip_list").StringArray()
	if err != nil {
		return nil, fmt.Errorf("format ip_list to []string err:%s\n", err.Error())
	}
	return ips, nil
}
