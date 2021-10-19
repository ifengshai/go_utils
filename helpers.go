package go_utils

import (
	"crypto/md5"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	mathRand "math/rand"
	"time"
)

// RandomNumber 生成随机数字
func RandomNumber(int int) int{
	mathRand.Seed(time.Now().UnixNano())
	return mathRand.Intn(int)
}

// UniqueId 生成32位md5 Guid字串
func UniqueId() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(cryptoRand.Reader, b); err != nil {
		return ""
	}
	h := md5.New()
	h.Write([]byte(base64.URLEncoding.EncodeToString(b)))
	return hex.EncodeToString(h.Sum(nil))
}

// InSlice 获取一个切片并在其中查找元素。如果找到它，它将返回它的密钥，否则它将返回-1和一个错误的bool。
func InSlice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}