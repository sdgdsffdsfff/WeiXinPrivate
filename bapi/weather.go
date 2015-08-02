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
	weather_apikey = "744a73341e3ab14cd6411bca70dbd5fd"
	weather_url    = "http://apis.baidu.com/apistore/weatherservice/cityname"
)

// {
// errNum: 0,
// errMsg: "success",
// retData: {
//    city: "北京", //城市
//    pinyin: "beijing", //城市拼音
//    citycode: "101010100",  //城市编码
//    date: "15-02-11", //日期
//    time: "11:00", //发布时间
//    postCode: "100000", //邮编
//    longitude: 116.391, //经度
//    latitude: 39.904, //维度
//    altitude: "33", //海拔
//    weather: "晴",  //天气情况
//    temp: "10", //气温
//    l_tmp: "-4", //最低气温
//    h_tmp: "10", //最高气温
//    WD: "无持续风向",	 //风向
//    WS: "微风(<10m/h)", //风力
//    sunrise: "07:12", //日出时间
//    sunset: "17:44" //日落时间
//   }
// }
type WeatherMessage struct {
	ErrNum  int         `json:"errNum"`
	ErrMsg  string      `json:"errMsg"`
	EetData WeatherData `json:"retData"`
}

type WeatherData struct {
	City      string  `json:"city"`
	Pinyin    string  `json:"pinyin"`
	Citycode  string  `json:"citycode"`
	Date      string  `json:"date"`
	Time      string  `json:"time"`
	PostCode  string  `json:"postCode"`
	Longitude float32 `json:"longitude"`
	Latitude  float32 `json:"latitude"`
	Altitude  string  `json:"altitude"`
	Weather   string  `json:"weather"`
	Temp      string  `json:"temp"`
	L_tmp     string  `json:"l_tmp"`
	H_tmp     string  `json:"h_tmp"`
	WD        string  `json:"WD"`
	WS        string  `json:"WS"`
	Sunrise   string  `json:"sunrise"`
	Sunset    string  `json:"sunset"`
}

func Weather(cityName string) (*WeatherMessage, error) {
	u, _ := url.Parse(weather_url)
	q := u.Query()
	q.Set("cityname", cityName)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("HTTP new request error. %s", err.Error())
	}
	client := &http.Client{}
	req.Header.Add("apikey", weather_apikey)
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

	var wm WeatherMessage
	json.Unmarshal(result, &wm)

	return &wm, nil
}

func (wm *WeatherMessage) ToString() string {
	data := wm.EetData
	return fmt.Sprintf("城市:%s %s\n发布时间:%s %s\n经纬:%0.3f %0.3f,海拔:%s\n天气情况:%s 气温:%s %s-%s\n风力:%s 风向:%s\n日出时间:%s 日落时间:%s",
		data.City, data.Pinyin, data.Date, data.Time, data.Longitude, data.Latitude, data.Altitude,
		data.Weather, data.Temp, data.L_tmp, data.H_tmp, data.WD, data.WS, data.Sunrise, data.Sunset)
}
