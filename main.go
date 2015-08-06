package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"

	"./bapi"
	"./cache"
	"./util"
	"./wapi"
)

const (
	token              = "RyanKatee"
	cache_file         = "./db.cache"
	defaultExpiration  = 60 * time.Minute
	cleanupInterval    = 10 * time.Second
	sleep_time         = 60 * time.Second
	key_access_token   = "Acess_Token"
	default_expires_in = "7200"
	main_menu          = `
	欢迎来到 [私人订制 Ryan_Katee] 
	命令菜单:
	1. "聊天" 或 "chat" + 内容 可以聊天哟!
	例: 聊天, 你好 或者 chat hello
	[powerd by 图灵机器人 http://www.tuling123.com/]
	2. "天气" 或 "weather" + 国内城市中文名 查询天气哟!
	例: 天气, 重庆 或者 weather 重庆
	3. "热门" 或者 "hot" 查看最新热门信息
	[powerd by 天行数据 http://api.huceo.com/]

	注意[命令]后可有一个空格或者逗号
	`
)

var (
	cacheServer *cache.WCache
	users       *wapi.UserList
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

func HandleMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Println("receive message...")

	reply := []byte("sorry...")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read req error:%s\n", err.Error())
		fmt.Fprintf(w, "success")
		return
	}
	msg, err := wapi.GetRequestBody(body)
	if err != nil {
		log.Printf("format message error:%s\n", err.Error())
		fmt.Fprintf(w, "success")
		return
	}

	// 文本消息回复逻辑
	content := `sorry...
		重复上次命令 
		或者
		回复: "菜单" 或 "M" 可查看命令菜单`
	if msg.MsgType == wapi.Message_type_event {
		key := msg.FromUserName + msg.CreateTime.String()
		_, found := cacheServer.Get(key)
		if !found {
			cacheServer.Add(key, 1, 30*time.Second)
			defer cacheServer.Delete(key)

			msg.Content = HandleEventMessage(msg.Event, msg.FromUserName)

			// 回复文本消息
			reply, err = wapi.GetTextResponseBody(msg)
			if err != nil {
				log.Printf("get send message error:%s\n", err.Error())
				fmt.Fprintf(w, "success")
				return
			}
		} else {
			log.Println("same FromUserName + CreateTime...")
			fmt.Fprintf(w, "success")
			return
		}
	} else if msg.MsgType == wapi.Message_type_text {
		key := strconv.Itoa(msg.MsgId)
		_, found := cacheServer.Get(key)
		if !found {
			mode := 1
			cacheServer.Add(key, 1, 30*time.Second)
			defer cacheServer.Delete(key)

			if msg.MsgType == wapi.Message_type_text {
				msg.Content, mode = HandleTextMessage(msg.Content)
			} else {
				msg.Content = content
			}

			if mode == 1 {
				// 回复文本消息
				reply, err = wapi.GetTextResponseBody(msg)
				if err != nil {
					log.Printf("get send message error:%s\n", err.Error())
					fmt.Fprintf(w, "success")
					return
				}
			} else if mode == 2 {
				// 回复图文消息
				ietms, err := wapi.GetNewsItemBody([]byte(msg.Content))
				if err != nil {
					log.Printf("get send message error:%s\n", err.Error())
					fmt.Fprintf(w, "success")
					return
				}
				reply, err = wapi.GetNewsResponseBody(msg, ietms)
				if err != nil {
					log.Printf("get send message error:%s\n", err.Error())
					fmt.Fprintf(w, "success")
					return
				}
			} else {
				// 回复文本消息
				reply, err = wapi.GetTextResponseBody(msg)
				if err != nil {
					log.Printf("get send message error:%s\n", err.Error())
					fmt.Fprintf(w, "success")
					return
				}
			}

		} else {
			log.Println("same MsgId...")
			fmt.Fprintf(w, "success")
			return
		}
	}

	fmt.Println(string(reply))
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprintf(w, string(reply))
}

func main() {
	// 多核设置
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 初始化cache
	cacheServer = &cache.WCache{}
	cacheServer.InitCache(cache_file, defaultExpiration, cleanupInterval)
	cacheServer.RunSaveMonitor(sleep_time)
	cacheServer.DeleteExpired()
	log.Println("init cache done...")

	// 初始化用户列表
	users = wapi.NewUserList()
	if list, found := cacheServer.Get("users"); found {
		switch list.(type) {
		case []string:
			users.Init(list.([]string))
		}
	} else {
		list := []string{}
		cacheServer.Add("users", list, cache.NoExpiration)
	}

	access_token, err := GetAccessToken()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("init access_token [%s] done...\n", access_token)

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/token", Token)
	router.GET("/test/getips", GetIPs)
	router.POST("/token", HandleMessage)

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

func FindInCacheThenDo(key string, foundFunc func(value interface{}) (string, int), notFoundFunc func() (string, int)) (string, int) {
	if cacheServer != nil {
		if val, found := cacheServer.Get(key); found {
			log.Println("found in cache...")
			return foundFunc(val)
		} else {
			log.Println("not found in cache...")
			return notFoundFunc()
		}
	} else {
		log.Println("cacheServer is nil...")
		return notFoundFunc()
	}
}

// 返回int 表示 回复模式 1:文本模式 2:图文模式
func HandleTextMessage(content string) (string, int) {
	if util.IsPre(content, []string{"菜单", "M"}) {
		return main_menu, 1
	} else if util.IsPre(content, []string{"聊天", "chat"}) {
		return FindInCacheThenDo(content,
			func(val interface{}) (string, int) {
				switch val.(type) {
				case string:
					return val.(string), 1
				default:
					return "found in cache, but parse to type [string] error.", 1
				}
			},
			func() (string, int) {
				info := util.Replace(content, []string{"聊天", "chat", " ", ",", "，"}, "")
				cm, err := bapi.Chat(info)
				if err != nil {
					log.Printf("chat tuning error. %s\n", err.Error())
					return "对不起,聊天功能故障...", 1
				} else {
					log.Printf("chat return code: %d text: %s\n", cm.Code, cm.Text)
					cacheServer.Add(content, cm.Text, 5*time.Minute)
					return cm.Text, 1
				}
			})
	} else if util.IsPre(content, []string{"天气", "weather"}) {
		return FindInCacheThenDo(content,
			func(val interface{}) (string, int) {
				switch val.(type) {
				case string:
					return val.(string), 1
				default:
					return "found in cache, but parse to type [string] error.", 1
				}
			},
			func() (string, int) {
				cityName := util.Replace(content, []string{"天气", "weather", " ", ",", "，"}, "")
				wm, err := bapi.Weather(cityName)
				if err != nil {
					log.Printf("chat tuning error. %s\n", err.Error())
					return "对不起,天气功能故障...", 1
				} else {
					log.Printf("weather return code: %d text: %s\n", wm.ErrNum, wm.ErrMsg)
					val := wm.ToString()
					cacheServer.Add(content, val, 5*time.Minute)
					return val, 1
				}
			})
	} else if util.IsPre(content, []string{"用户", "user"}) {
		return FindInCacheThenDo(content,
			func(val interface{}) (string, int) {
				switch val.(type) {
				case string:
					return val.(string), 1
				default:
					return "found in cache, but parse to type [string] error.", 1
				}
			},
			func() (string, int) {
				//message := util.Replace(content, []string{"广播", "send", " ", ",", "，"}, "")
				if val, found := cacheServer.Get("users"); found {
					log.Println("found users in cache...")
					switch val.(type) {
					case []string:
						us := val.([]string)
						var list = strconv.Itoa(len(us))
						for _, user := range us {
							list += "\n" + user
						}
						return list, 1
					default:
						return "no user", 1
					}
				} else {
					return "no user", 1
				}
			})
	} else if util.IsPre(content, []string{"热门", "hot"}) {
		return FindInCacheThenDo(content,
			func(val interface{}) (string, int) {
				switch val.(type) {
				case string:
					return val.(string), 1
				default:
					return "found in cache, but parse to type [string] error.", 1
				}
			},
			func() (string, int) {
				word := util.Replace(content, []string{"热门", "hot", " ", ",", "，"}, "")
				wm, err := bapi.WxHot(word)
				if err != nil {
					log.Printf("chat tuning error. %s\n", err.Error())
					return "对不起,热门功能故障...", 1
				} else {
					log.Printf("wxhot return code: %d text: %s\n", wm.Code, wm.Msg)
					if wm.Code == 200 {
						val, err := wm.ToString()
						if err != nil {
							log.Printf("wxhot format error. %s\n", err.Error())
							return "对不起,热门功能故障...", 1
						}
						cacheServer.Add(content, string(val), 3*time.Hour)
						return string(val), 2
					} else {
						val := wm.Msg
						cacheServer.Add(content, val, 3*time.Hour)
						return val, 1
					}
				}
			})
	} else {
		return content, 1
	}
}

func HandleEventMessage(event string, openid string) string {
	if event == wapi.Message_type_event_sub {
		if !users.IsExist(openid) {
			users.Add(openid)
		}
		defer cacheServer.Set("users", users.GetAll(), cache.NoExpiration)
		return main_menu
	} else if event == wapi.Message_type_event_unsub {
		users.Remove(openid)
		defer cacheServer.Set("users", users.GetAll(), cache.NoExpiration)
		return "success"
	} else {
		return "unkown event " + event
	}
}
