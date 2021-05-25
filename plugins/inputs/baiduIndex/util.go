package baiduIndex

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

func Run(method, path, body string, headers map[string]string) (int, string) {
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Content-Type"] = "application/json"

	method = strings.ToUpper(method)
	if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
		l.Error("Unsupported HTTP Method: " + method)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	req, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		l.Error(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Error(err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		l.Error(err)
	}

	respStatusCode := resp.StatusCode
	respData := string(respBody)

	return respStatusCode, respData
}

func Get(path string, cookie string) (int, string) {
	headers := map[string]string{
		"Cookie":           cookie,
		"Host":             "index.baidu.com",
		"Connection":       "keep-alive",
		"X-Requested-With": "XMLHttpRequest",
		"User-Agent":       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	}

	return Run("GET", path, "", headers)
}

func SplitWord(words []string) [][]string {
	var list [][]string
	var item = []string{}
	for idx, word := range words {
		if (idx+1)%5 == 0 {
			list = append(list, item)
		} else {
			item = append(item, word)
		}
	}

	return list
}

func getKey(uniqid, cookie string) string {
	url := fmt.Sprintf("http://index.baidu.com/Interface/api/ptbk?uniqid=%s", uniqid)
	_, resp := Get(url, cookie)

	key := gjson.Parse(resp).Get("data").String()
	return key
}

func decrypt(key, data string) string {
	a := key
	i := data

	n := map[string]string{}
	s := []string{}

	num := len(a) / 2
	for o := 0; o < num; o++ {
		idx := string(a[o])
		n[idx] = string(a[num+o])
	}

	for r := 0; r < len(data); r++ {
		key := string(i[r])
		s = append(s, n[key])
	}
	res := strings.Join(s, "")

	if res == "" {
		res = "0"
	}
	return res
}

func ConvertToFloat(str string) float64 {
	value, _ := strconv.ParseFloat(str, 64)
	return value
}

func ConvertToNum(str string) int64 {
	num, _ := strconv.ParseInt(str, 10, 64)

	return num
}
