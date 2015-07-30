package wapi

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	//"github.com/bitly/go-simplejson"
)

const (
	appid             = "wxafc2e7724ecc19e9"
	appsecret         = "e0c95308a69c57355064aaf8de85c59d"
	access_token_url  = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET"
	server_ips_url    = "https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=ACCESS_TOKEN"
	message_type_text = "text"
)

// {"errcode":40013,"errmsg":"invalid appid"}
type WError struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func (e *WError) transToError() error {
	// TODO 转义全局错误代码
	return fmt.Errorf("weixin errcode:%s errmsg:%s", e.Errcode, e.Errmsg)
}

// {"access_token":"ACCESS_TOKEN","expires_in":7200}
type AccessToken struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
}

// {"ip_list":["127.0.0.1","127.0.0.1"]}
type ServerIP struct {
	Ip_list []string `json:"ip_list"`
}

// <xml>
// <ToUserName><![CDATA[toUser]]></ToUserName>
// <CreateTime>1348831860</CreateTime>
// <MsgType><![CDATA[text]]></MsgType>
// <Content><![CDATA[this is a test]]></Content>
// <MsgId>1234567890123456</MsgId>
// </xml>
type TextRequestBody struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string
	FromUserName string
	CreateTime   time.Duration
	MsgType      string
	Content      string
	MsgId        int
}

// <xml>
// <ToUserName><![CDATA[toUser]]></ToUserName>
// <FromUserName><![CDATA[fromUser]]></FromUserName>
// <CreateTime>12345678</CreateTime>
// <MsgType><![CDATA[text]]></MsgType>
// <Content><![CDATA[你好]]></Content>
// </xml>
type TextResponseBody struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   CDATAText
	FromUserName CDATAText
	CreateTime   time.Duration
	MsgType      CDATAText
	Content      CDATAText
}

type CDATAText struct {
	Text string `xml:",innerxml"`
}

func value2CDATA(v string) CDATAText {
	return CDATAText{"<![CDATA[" + v + "]]>"}
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

	var token AccessToken
	json.Unmarshal(body, &token)

	if token.Access_token != "" && token.Expires_in != 0 {
		return &token, nil
	} else {
		var werr WError
		json.Unmarshal(body, werr)

		if werr.Errcode != 0 && werr.Errmsg != "" {
			return nil, fmt.Errorf("can not format [%s].", string(body))
		} else {
			return nil, werr.transToError()
		}
	}
}

func GetWeixinServerIP(access_token string) (*ServerIP, error) {
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

	var ips ServerIP
	json.Unmarshal(body, &ips)

	if len(ips.Ip_list) != 0 {
		return &ips, nil
	} else {
		var werr WError
		json.Unmarshal(body, &werr)

		if werr.Errcode != 0 && werr.Errmsg != "" {
			return nil, fmt.Errorf("can not format [%s].", string(body))
		} else {
			return nil, werr.transToError()
		}
	}
}

func GetTextRequestBody(body []byte) (*TextRequestBody, error) {
	var msg TextRequestBody
	xml.Unmarshal(body, &msg)

	if msg.FromUserName != "" && msg.ToUserName != "" && msg.MsgType != "" && msg.Content != "" {
		return &msg, nil
	} else {
		return nil, fmt.Errorf("xml Unmarshal error.")
	}
}

func GetTextResponseBody(text *TextRequestBody) ([]byte, error) {
	content := `sorry...
		回复: "菜单"或"M" 可查看命令菜单`
	if text != nil {
		if text.MsgType == message_type_text {
			if text.Content == "菜单" || text.Content == "M" {
				content = `欢迎来到 [私人订制 Ryan_Katee] 
				命令菜单:
				1. asd [敬请期待]
				2. asd [敬请期待]
				`
			}
		}
	}

	respMsg := &TextResponseBody{}
	respMsg.FromUserName = value2CDATA(text.ToUserName)
	respMsg.ToUserName = value2CDATA(text.FromUserName)
	respMsg.MsgType = value2CDATA(text.MsgType)
	respMsg.CreateTime = time.Duration(time.Now().Unix())
	respMsg.Content = value2CDATA(content)
	return xml.Marshal(&respMsg)
}
