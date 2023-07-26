package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
)

func M5(src string) string {
	m := md5.New()
	m.Write([]byte(src))
	return fmt.Sprintf("%x", m.Sum(nil))
}

func HMAC(src, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(src))
	return fmt.Sprintf("%x", h.Sum(nil))
}
