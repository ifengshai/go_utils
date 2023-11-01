package go_utils

import (
	"bytes"
	"crypto/md5"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	mathRand "math/rand"
	"net/http"
	"os"
	"time"
)

// RandomNumber 生成随机数字
func RandomNumber(int int) int {
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

// PathExists @author: [piexlmax](https://github.com/piexlmax)
// @function: PathExists
// @description: 文件目录是否存在
// @param: path string
// @return: bool, error
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateDir @author: [piexlmax](https://github.com/piexlmax)
// @function: CreateDir
// @description: 批量创建文件夹
// @param: dirs ...string
// @return: err error
func CreateDir(dirs ...string) (err error) {
	for _, v := range dirs {
		exist, err := PathExists(v)
		if err != nil {
			return err
		}
		if !exist {
			fmt.Println("create directory" + v)
			if err := os.MkdirAll(v, os.ModePerm); err != nil {
				fmt.Println("create directory"+v, zap.Any(" error:", err))
				return err
			}
		}
	}
	return err
}

// SendHTTPRequest 发送HTTP请求并返回响应数据
// url: 请求的URL地址
// method: 请求的方法（GET、POST等）
// headers: 请求头部信息
// body: 请求体数据
// timeout: 请求超时时间
// retryCount: 请求重试次数
// retryInterval: 请求重试间隔时间
// alarm: 是否开启请求失败告警
// ([]byte, error): 返回响应的字节数组和可能的错误
func SendHTTPRequest(url string, method string, headers map[string]string, body []byte,
	timeout time.Duration, retryCount int, retryInterval time.Duration, alarm bool) ([]byte, error) {

	var err error
	var responseBody []byte

	for i := 0; i < retryCount+1; i++ {
		req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
		if err != nil {
			if i == retryCount && alarm {
				fmt.Printf("新建请求出错：%s\n", err.Error())
			}
			continue
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		client := &http.Client{Timeout: timeout}
		resp, err := client.Do(req)
		if err != nil {
			if i == retryCount && alarm {
				fmt.Printf("第 %d 次请求失败：%s\n", i+1, err.Error())
			}
			time.Sleep(retryInterval)
			continue
		}
		defer resp.Body.Close()

		responseBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			if i == retryCount && alarm {
				fmt.Printf("第 %d 次请求失败：读取响应出错：%s\n", i+1, err.Error())
			}
			time.Sleep(retryInterval)
			continue
		}

		break
	}

	if err != nil {
		return nil, err
	}

	return responseBody, nil
}
