package wapi

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	fromUserName             = "gh_4c204e94e7a0"
	appid                    = "wxafc2e7724ecc19e9"
	appsecret                = "e0c95308a69c57355064aaf8de85c59d"
	access_token_url         = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=APPID&secret=APPSECRET"
	server_ips_url           = "https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=ACCESS_TOKEN"
	Message_type_text        = "text"
	Message_type_event       = "event"
	Message_type_event_sub   = "subscribe"
	Message_type_event_unsub = "unsubscribe"
	Message_type_news        = "news"
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

// {"ip_list":["127.0.0.1","127.0.0.1"]}
type ServerIP struct {
	Ip_list []string `json:"ip_list"`
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

// <xml>
// <ToUserName><![CDATA[toUser]]></ToUserName>
// <CreateTime>1348831860</CreateTime>
// <MsgType><![CDATA[text]]></MsgType>
// <Content><![CDATA[this is a test]]></Content>
// <MsgId>1234567890123456</MsgId>
// <Event><![CDATA[subscribe]]></Event>
// </xml>
type RequestBody struct {
	XMLName      xml.Name      `xml:"xml"`
	ToUserName   string        `xml:"ToUserName"`
	FromUserName string        `xml:"FromUserName"`
	CreateTime   time.Duration `xml:"CreateTime"`
	MsgType      string        `xml:"MsgType"`
	Content      string        `xml:"Content"`
	MsgId        int           `xml:"MsgId"`
	Event        string        `xml:"Event"`
	EventKey     string        `xml:"EventKey"`
}

func GetRequestBody(body []byte) (*RequestBody, error) {
	log.Printf("request body %s\n", string(body))
	var msg RequestBody
	xml.Unmarshal(body, &msg)

	if msg.FromUserName != "" && msg.ToUserName != "" && msg.MsgType != "" {
		return &msg, nil
	} else {
		return nil, fmt.Errorf("xml Unmarshal error.")
	}
}

type CDATAText struct {
	Text string `xml:",innerxml"`
}

func value2CDATA(v string) CDATAText {
	return CDATAText{"<![CDATA[" + v + "]]>"}
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

func GetTextResponseBody(text *RequestBody) ([]byte, error) {
	respMsg := &TextResponseBody{}
	respMsg.FromUserName = value2CDATA(text.ToUserName)
	respMsg.ToUserName = value2CDATA(text.FromUserName)
	respMsg.MsgType = value2CDATA(Message_type_text)
	respMsg.CreateTime = time.Duration(time.Now().Unix())
	respMsg.Content = value2CDATA(text.Content)
	return xml.Marshal(&respMsg)
}

// <xml>
// <ToUserName><![CDATA[toUser]]></ToUserName>
// <FromUserName><![CDATA[fromUser]]></FromUserName>
// <CreateTime>12345678</CreateTime>
// <MsgType><![CDATA[news]]></MsgType>
// <ArticleCount>2</ArticleCount>
// <Articles>
// <item>
// <Title><![CDATA[title1]]></Title>
// <Description><![CDATA[description1]]></Description>
// <PicUrl><![CDATA[picurl]]></PicUrl>
// <Url><![CDATA[url]]></Url>
// </item>
// <item>
// <Title><![CDATA[title]]></Title>
// <Description><![CDATA[description]]></Description>
// <PicUrl><![CDATA[picurl]]></PicUrl>
// <Url><![CDATA[url]]></Url>
// </item>
// </Articles>
// </xml>
type NewsResponseBody struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   CDATAText
	FromUserName CDATAText
	CreateTime   time.Duration
	MsgType      CDATAText
	ArticleCount int
	Articles     ArticlesBody
}

type ArticlesBody struct {
	XMLName xml.Name   `xml:"Articles"`
	Item    []ItemBody `xml:"item"`
}

type ItemBody struct {
	XMLName     xml.Name
	Title       CDATAText
	Description CDATAText
	PicUrl      CDATAText
	Url         CDATAText
}

type Item struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PicUrl      string `json:"picUrl"`
	Url         string `json:"url"`
}

func GetNewsResponseBody(text *RequestBody, items []ItemBody) ([]byte, error) {
	respMsg := &NewsResponseBody{}
	respMsg.FromUserName = value2CDATA(text.ToUserName)
	respMsg.ToUserName = value2CDATA(text.FromUserName)
	respMsg.MsgType = value2CDATA(Message_type_news)
	respMsg.CreateTime = time.Duration(time.Now().Unix())
	respMsg.ArticleCount = len(items)
	articles := ArticlesBody{}
	articles.Item = items
	respMsg.Articles = articles
	return xml.Marshal(&respMsg)
}

func GetNewsItemBody(data []byte) ([]ItemBody, error) {
	var itemb []ItemBody
	var items []Item
	err := json.Unmarshal(data, &items)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		ib := ItemBody{}
		ib.Title = value2CDATA(item.Title)
		ib.Description = value2CDATA(item.Description)
		ib.PicUrl = value2CDATA(item.PicUrl)
		ib.Url = value2CDATA(item.Url)
		itemb = append(itemb, ib)
	}
	return itemb, err
}

type UserList struct {
	openIds []string
	locker  *sync.Mutex
}

func (u *UserList) Init(ids []string) {
	u.locker.Lock()
	u.openIds = ids
	u.locker.Unlock()
}

func (u *UserList) Add(openId string) {
	u.locker.Lock()
	u.openIds = append(u.openIds, openId)
	u.locker.Unlock()
}

func (u *UserList) Remove(openId string) {
	u.locker.Lock()
	for i, v := range u.openIds {
		if v == openId {
			u.openIds = append(u.openIds[:i], u.openIds[i+1:]...)
			break
		}
	}
	u.locker.Unlock()
}

func (u *UserList) GetAll() []string {
	return u.openIds
}

func (u *UserList) IsExist(openId string) bool {
	u.locker.Lock()
	for _, v := range u.openIds {
		if v == openId {
			return true
		}
	}
	u.locker.Unlock()
	return false
}

func NewUserList() *UserList {
	return &UserList{[]string{}, &sync.Mutex{}}
}
