package util

import (
	"crypto/sha1"
	"fmt"
	"io"
)

/**
 * 将data已sha1加密
 */
func CryptoSHA1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}
