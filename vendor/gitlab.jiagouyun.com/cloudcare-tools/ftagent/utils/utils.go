package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GinGetArg(c *gin.Context, headerName, argName string) string {
	v := c.Request.Header.Get(headerName)
	if v == "" {
		v = c.Query(argName)
	}

	return v
}

func ContainsValue(target string, source []string) bool {
	for i := 0; i < len(source); i++ {
		if source[i] == target {
			return true
		}
	}
	return false
}

func ReadCompressed(body *bytes.Reader, isGzip bool) (data []byte, err error) {

	if isGzip {
		reader, err_ := gzip.NewReader(body)
		if err_ != nil {
			log.Printf("[error] %s", err_.Error())
			return nil, err_
		}

		data, err = ioutil.ReadAll(reader)
		if err != nil {
			log.Printf("[error] %s", err.Error())
			return
		}

	} else {
		data, err = ioutil.ReadAll(body)
		if err != nil {
			log.Printf("[error] %s", err.Error())
			return
		}
	}

	return
}

// 签名算法
//  content: 上传的数据 body
//  contentType: 上传数据格式  如“application/json”
//  dateStr: header "Date"
//  key: header "X-Team-ID"
//  method: http request method ,eg PUT，GET，POST，HEAD，DELETE
//  skVal: Access key secret
func Signature(content []byte, contentType, dateStr, key, method, skVal string) (string, string) {
	var orignalStr string

	mac := hmac.New(sha1.New, []byte(skVal))

	h := md5.New()
	h.Write(content)
	cipherStr := h.Sum(nil)

	orignalStr = method + "\n" + hex.EncodeToString(cipherStr) + "\n" + contentType + "\n" + dateStr + "\n" + key

	mac.Write([]byte(orignalStr))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return sig, orignalStr
}

func SetSysLimit() error {
	// var limit syscall.Rlimit
	// err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit)
	// if err != nil {
	// 	return err
	// }

	// limit.Max = 999999
	// limit.Cur = 999999
	// return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit)
	return nil
}

func FormatTimeStamps(n int64) (int64, error) {
	s := strconv.FormatInt(n, 10)
	switch len(s) {
	case 10:
		return n, nil
	case 13:
		return n / 1000, nil
	case 16:
		return n / 1000000, nil
	case 19:
		return n / 1000000000, nil
	default:
		return -1, ErrTimeStamp
	}
}
