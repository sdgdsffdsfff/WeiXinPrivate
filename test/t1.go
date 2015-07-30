package main

import (
	//"encoding/json"
	"encoding/xml"
	"fmt"
	"time"
)

type TextRequestBody struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string
	FromUserName string
	CreateTime   time.Duration
	MsgType      string
	Content      string
	MsgId        int
}

type ToUserName struct {
}

func main() {
	x := `<xml>
 <ToUserName><![CDATA[toUser]]></ToUserName>
 <FromUserName><![CDATA[fromUser]]></FromUserName> 
 <CreateTime>1348831860</CreateTime>
 <MsgType><![CDATA[text]]></MsgType>
 <Content><![CDATA[this is a test]]></Content>
 <MsgId>1234567890123456</MsgId>
 </xml>`
	var v TextRequestBody
	xml.Unmarshal([]byte(x), &v)
	fmt.Println(v)
	s, _ := xml.Marshal(v)
	fmt.Println(string(s))
	fmt.Println(time.Now().Format("02150405"))
}
