package bapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	chat_apikey = "744a73341e3ab14cd6411bca70dbd5fd"
	chat_url    = "http://apis.baidu.com/turing/turing/turing"
	chat_key    = "879a6cb3afb84dbf4fc84a1df2ab7319"
	chat_userid = "eb2edb736"
)

type ChatMessage struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func Chat(info string) (*ChatMessage, error) {
	u, _ := url.Parse(chat_url)
	q := u.Query()
	q.Set("key", chat_key)
	q.Set("info", info)
	q.Set("userid", chat_userid)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("HTTP new request error. %s", err.Error())
	}
	client := &http.Client{}
	req.Header.Add("apikey", chat_apikey)
	res, err := client.Do(req)
	defer res.Body.Close()

	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("HTTP do get error. %s", err.Error())
	}

	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("HTTP read response error. %s", err.Error())
	}
	log.Printf("HTTP STATUS %d %s\n%s\n", res.StatusCode, res.Status, result)

	var cm ChatMessage
	json.Unmarshal(result, &cm)

	return &cm, nil
}
