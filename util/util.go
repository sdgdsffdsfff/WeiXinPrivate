package util

import (
	"crypto/sha1"
	"fmt"
	"io"
	"strings"
)

// 将data已sha1加密
func CryptoSHA1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

// 如果s已pre中任一字符串开头则替换成to
func Replace(s string, pre []string, to string) string {
	for _, v := range pre {
		i := strings.Index(s, v)
		if i == 0 {
			return strings.Replace(s, v, to, 1)
		}
	}
	return s
}

// 查看s是否以pre中任一字符串开头是ture
func IsPre(s string, pre []string) bool {
	for _, v := range pre {
		i := strings.Index(s, v)
		if i == 0 {
			return true
		}
	}
	return false
}
