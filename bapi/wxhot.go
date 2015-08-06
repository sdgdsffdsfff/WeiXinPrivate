package bapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const (
	wxhot_apikey = "744a73341e3ab14cd6411bca70dbd5fd"
	wxhot_url    = "http://apis.baidu.com/txapi/weixin/wxhot"
	wxhot_rand   = 1
	wxhot_count  = 10
	wxhot_page   = 1
)

// {
//     "0": {
//         "title": "玻璃心的人都会喜欢这些东西",
//         "description": "一条",
//         "picUrl": "http://zxpic.gtimg.com/infonew/0/wechat_pics_-551403.jpg/640",
//         "url": "http://mp.weixin.qq.com/s?__biz=MjM5MDI5OTkyOA==&idx=1&mid=222639330&sn=b7b0425cde58aea059d9376da0d40444&qb_mtt_show_type=1"
//     },
//     "1": {
//         "title": "四个版本的《当你老了》，总有一个让你泪流满面",
//         "description": "水木文摘",
//         "picUrl": "http://zxpic.gtimg.com/infonew/0/wechat_pics_-536525.jpg/640",
//         "url": "http://mp.weixin.qq.com/s?__biz=MjM5ODExNDM0Mg==&idx=3&mid=212262164&sn=a9ae8369a35ffe4b0fd64ee995830f1b&qb_mtt_show_type=1"
//     },
//     "2": {
//         "title": "火爆广州的暖男海鲜外卖，来北京啦！",
//         "description": "美味食尚",
//         "picUrl": "http://zxpic.gtimg.com/infonew/0/wechat_pics_-540930.jpg/640",
//         "url": "http://mp.weixin.qq.com/s?__biz=MzA4NzAwODkzMQ==&idx=1&mid=228554902&sn=44b59b787a23ff0aab5fdb0794eefcaa&qb_mtt_show_type=1"
//     },
//     "3": {
//         "title": "【会计实务】如何编制应付职工薪酬的会计分录",
//         "description": "会计网",
//         "picUrl": "http://zxpic.gtimg.com/infonew/0/wechat_pics_-532671.jpg/640",
//         "url": "http://mp.weixin.qq.com/s?__biz=MjM5NjAzNzE2MA==&idx=2&mid=210824204&sn=f9cddf788f8dab4f0d752b270fefa9a6&qb_mtt_show_type=1"
//     },
//     "4": {
//         "title": "解密德尼罗不败战神之谜《致命对决》成就多项突破战绩",
//         "description": "热荐电影",
//         "picUrl": "http://zxpic.gtimg.com/infonew/0/wechat_pics_-546698.jpg/640",
//         "url": "http://mp.weixin.qq.com/s?__biz=MjM5NTAzNzE4MA==&idx=3&mid=206584799&sn=58099643e4b09ee2d40c355d6aa30d8e&qb_mtt_show_type=1"
//     },
//     "code": 200,
//     "msg": "ok"
// }
type WxHotMessage struct {
	M0   WxHotItem `json:"0"`
	M1   WxHotItem `json:"1"`
	M2   WxHotItem `json:"2"`
	M3   WxHotItem `json:"3"`
	M4   WxHotItem `json:"4"`
	M5   WxHotItem `json:"5"`
	M6   WxHotItem `json:"6"`
	M7   WxHotItem `json:"7"`
	M8   WxHotItem `json:"8"`
	M9   WxHotItem `json:"9"`
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
}

type WxHotItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PicUrl      string `json:"picUrl"`
	Url         string `json:"url"`
}

func WxHot(word string) (*WxHotMessage, error) {
	u, _ := url.Parse(wxhot_url)
	q := u.Query()
	q.Set("num", strconv.Itoa(wxhot_count))
	//q.Set("rand", strconv.Itoa(wxhot_rand))
	//q.Set("page", strconv.Itoa(wxhot_page))
	if word != "" {
		q.Set("word", word)
	}
	u.RawQuery = q.Encode()
	log.Printf("URL:%s\n", u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("HTTP new request error. %s", err.Error())
	}
	client := &http.Client{}
	req.Header.Add("apikey", wxhot_apikey)
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

	var wm WxHotMessage
	json.Unmarshal(result, &wm)

	return &wm, nil
}

func (w *WxHotMessage) ToString() ([]byte, error) {
	arr := []WxHotItem{}
	if w.M0.Title != "" {
		arr = append(arr, w.M0)
	}
	if w.M1.Title != "" {
		arr = append(arr, w.M1)
	}
	if w.M2.Title != "" {
		arr = append(arr, w.M2)
	}
	if w.M3.Title != "" {
		arr = append(arr, w.M3)
	}
	if w.M4.Title != "" {
		arr = append(arr, w.M4)
	}
	if w.M5.Title != "" {
		arr = append(arr, w.M5)
	}
	if w.M6.Title != "" {
		arr = append(arr, w.M6)
	}
	if w.M7.Title != "" {
		arr = append(arr, w.M7)
	}
	if w.M8.Title != "" {
		arr = append(arr, w.M8)
	}
	if w.M9.Title != "" {
		arr = append(arr, w.M9)
	}
	return json.Marshal(arr)
}
