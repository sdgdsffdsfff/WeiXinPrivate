package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"

	"./cache"
	"./util"
	"./wapi"
)

const (
	token              = "RyanKatee"
	cache_file         = "./db.cache"
	defaultExpiration  = 60 * time.Minute
	cleanupInterval    = 10 * time.Second
	sleep_time         = 10 * time.Second
	key_access_token   = "Acess_Token"
	default_expires_in = "7200"
)

var (
	cacheServer *cache.WCache
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

/**
 * 微信订阅号打开开发者模式时需要验证的Token方法
 */
func Token(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	signature := r.Form.Get("signature")
	timestamp := r.Form.Get("timestamp")
	nonce := r.Form.Get("nonce")
	echostr := r.Form.Get("echostr")
	fmt.Println("signature", signature)
	fmt.Println("timestamp", timestamp)
	fmt.Println("nonce", nonce)
	fmt.Println("echostr", echostr)

	arr := []string{token, timestamp, nonce}
	fmt.Println(arr)
	sort.Strings(arr)
	fmt.Println(arr)
	sha1 := util.CryptoSHA1(arr[0] + arr[1] + arr[2])
	fmt.Println("sha1", sha1)
	if signature == sha1 {
		fmt.Println("equal...")
		fmt.Fprintf(w, "%s", echostr)
	} else {
		fmt.Printf("%s", "not validate")
		fmt.Fprintf(w, "success")
	}
}

func GetIPs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	token, found := cacheServer.Get(key_access_token)
	if !found {
		fmt.Printf("not get access_token from cache.")
		fmt.Fprintf(w, "success")
		return
	}
	switch token.(type) {
	case string:
		ips, err := wapi.GetWeixinServerIP(token.(string))
		if err != nil {
			log.Printf("get weixin server ip err. %s", err.Error())
			fmt.Fprintf(w, "success")
			return
		}
		fmt.Fprintf(w, "get weixin server ip : %s", ips.Ip_list)
	default:
		fmt.Printf("access_token from cache is not type string.")
		fmt.Fprintf(w, "success")
	}
}

func TextMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("receive message...")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read req error:%s\n", err.Error())
		fmt.Fprintf(w, "success")
		return
	}
	msg, err := wapi.GetTextRequestBody(body)
	if err != nil {
		log.Printf("format message error:%s\n", err.Error())
		fmt.Fprintf(w, "success")
		return
	}
	_, found := cacheServer.Get(strconv.Itoa(msg.MsgId))
	if !found {
		cacheServer.Add(strconv.Itoa(msg.MsgId), 1, 30*time.Second)
		defer cacheServer.Delete(strconv.Itoa(msg.MsgId))
		// 回复消息
		content, err := wapi.GetTextResponseBody(msg)
		if err != nil {
			log.Printf("get send message error:%s\n", err.Error())
			fmt.Fprintf(w, "success")
			return
		}
		fmt.Println(string(content))
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprintf(w, string(content))
	} else {
		fmt.Printf("same msgId...")
	}
}

func main() {
	// 初始化cache
	cacheServer = &cache.WCache{}
	cacheServer.InitCache(cache_file, defaultExpiration, cleanupInterval)
	cacheServer.RunSaveMonitor(sleep_time)
	cacheServer.DeleteExpired()
	log.Println("init cache done...")

	access_token, err := GetAccessToken()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("init access_token [%s] done...\n", access_token)

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/token", Token)
	router.GET("/test/getips", GetIPs)
	router.POST("/token", TextMessage)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func GetAccessToken() (string, error) {
	if cacheServer != nil {
		tokenS, found := cacheServer.Get(key_access_token)
		if found {
			switch tokenS.(type) {
			case string:
				return tokenS.(string), nil
			default:
				return "", fmt.Errorf("access_token from cache is not type string.\n")
			}
		}
		token, err := wapi.GetAcessToken()
		if err != nil {
			return "", fmt.Errorf("wapi GetAcessToken error. %s\n", err.Error())
		}
		expires_in, err := time.ParseDuration(strconv.Itoa(token.Expires_in) + "s")
		if err != nil {
			log.Panicf("format expires [%s] err:%s\n", token.Expires_in, err.Error())
			log.Printf("use default expires_in %s", default_expires_in)
			expires_in, err = time.ParseDuration(default_expires_in + "s")
			if err != nil {
				return "", fmt.Errorf("format default expires [%s] err:%s\n", token.Expires_in, err.Error())
			}
		}
		cacheServer.Add(key_access_token, token.Access_token, expires_in)
		return token.Access_token, nil
	}
	return "", fmt.Errorf("cacheServer is nil.\n")
}
