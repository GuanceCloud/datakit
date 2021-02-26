package utils

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var durationRE = regexp.MustCompile("^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?(([0-9]+)us)?(([0-9]+)ns)?$")

func ParseDuration(s string) (time.Duration, error) {
	switch s {
	case "0":
		return 0, nil
	case "":
		return 0, fmt.Errorf("empty duration string")
	}

	m := durationRE.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid duration string: %q", s)
	}

	var du time.Duration
	f := func(pos int, mult time.Duration) {
		if m[pos] == "" {
			return
		}

		n, _ := strconv.Atoi(m[pos])
		d := time.Duration(n)
		du += d * mult
	}

	f(2, 60*60*24*365*time.Second) // y
	f(4, 60*60*24*7*time.Second)   // w
	f(6, 60*60*24*time.Second)     // d
	f(8, 60*60*time.Second)        // h
	f(10, 60*time.Second)          // m
	f(12, time.Second)             // s
	f(14, time.Millisecond)        // ms
	f(16, time.Microsecond)        // us
	f(18, time.Nanosecond)         // ns
	return time.Duration(du), nil
}

func FormatRequest(r *http.Request) string {
	var request []string

	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers

	for name, headers := range r.Header {
		// name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		request = append(request, "\n")
	}

	// Return the request as a string
	return strings.Join(request, "\n")
}

var (
	letterBytes   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = uint(6)              // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits
)

func CreateRandomString(n int) string {

	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & int64(letterIdxMask)); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func ContainsValue(target string, source []string) bool {
	for i := 0; i < len(source); i++ {
		if source[i] == target {
			return true
		}
	}
	return false
}

func ContainsIntValue(target int, source []int) bool {
	for i := 0; i < len(source); i++ {
		if source[i] == target {
			return true
		}
	}
	return false
}

func RandStringBytesRmndr(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func StringsNotExist(source []string, des map[string]interface{}) []string {
	var res []string
	for _, s := range source {
		if _, ok := des[s]; !ok {
			res = append(res, s)
		}
	}
	return res
}

func CktableInfoRedisKey(uuid, tablename string) string {
	return `cktabinfo:` + uuid + `_` + tablename
}

func CountMd5Value(source interface{}) (string, error) {
	datas, err := json.Marshal(source)
	if err != nil {
		return "", err
	}

	md5Cur := fmt.Sprintf("%x", md5.Sum(datas)) //将[]byte转成16进制

	return md5Cur, nil
}

func CountMd5Map(source map[string]interface{}) (string, error) {
	keys := []string{}
	for k := range source {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	digest := md5.New()
	for _, k := range keys {
		digest.Write([]byte(k))
		data, err := json.Marshal(source[k])
		if err != nil {
			return "", err
		}
		digest.Write(data)
	}

	md5Cur := fmt.Sprintf("%x", digest.Sum(nil))
	return md5Cur, nil
}

func CountMd5MapString(source map[string]string) (string, error) {
	keys := []string{}
	for k := range source {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	digest := md5.New()
	for _, k := range keys {
		digest.Write([]byte(k))
		data, err := json.Marshal(source[k])
		if err != nil {
			return "", err
		}
		digest.Write(data)
	}

	md5Cur := fmt.Sprintf("%x", digest.Sum(nil))
	return md5Cur, nil
}

// 判断文件夹是否存在
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

func AppendList(key string, source []string) ([]string, error) {

	var err error
	if !ContainsValue(key, source) {
		source = append(source, key)
		return source, nil
	}

	index := 0
	ms := strings.Split(key, `_`)
	if len(ms) == 2 {
		index, err = strconv.Atoi(ms[1])
		if err != nil {
			return nil, err
		}
	}

	index++
	keyNew := fmt.Sprintf(`%s_%d`, key, index)

	return AppendList(keyNew, source)

}

type KeepZero float64

func (f KeepZero) MarshalJSON() ([]byte, error) {
	if float64(f) == math.Trunc(float64(f)) {
		return []byte(fmt.Sprintf("%.1f", f)), nil
	}
	return []byte(fmt.Sprintf("%f", f)), nil
}

func KeepZeroMap(source map[string]interface{}) map[string]interface{} {
	for k, v := range source {
		switch v.(type) {
		case float64:
			source[k] = KeepZero(v.(float64))
		case float32:
			source[k] = KeepZero(v.(float32))
		case map[string]interface{}:
			source[k] = KeepZeroMap(v.(map[string]interface{}))
		default:

		}
	}

	return source
}
