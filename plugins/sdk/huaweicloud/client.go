package huaweicloud

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

type HWClient struct {
	ak        string
	sk        string
	endpoint  string
	projectid string
	logger    *logger.Logger
}

func NewHWClient(ak, sk, endpoint, projectid string, logger *logger.Logger) *HWClient {
	return &HWClient{
		ak:        ak,
		sk:        sk,
		endpoint:  endpoint,
		projectid: projectid,
		logger:    logger,
	}
}

func (c *HWClient) Request(method string, resPath string, querys map[string]string, body []byte) ([]byte, error) {

	canonicalRequest := method + "\n"

	canonicalURI := resPath + "/"
	canonicalRequest += canonicalURI + "\n"

	queryString := c.genCanonicalQuery(querys)
	canonicalRequest += queryString + "\n"

	now := time.Now().UTC()
	xdate := fmt.Sprintf("%d%02d%02dT%02d%02d%02dZ", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

	headers := map[string]string{
		`X-Sdk-Date`:   xdate,
		`Content-Type`: `application/json`,
		`Host`:         c.endpoint,
	}

	canonicalRequest += c.genCanonicalHeaders(headers)

	signedHeaders := []string{
		`X-Sdk-Date`,
		`Content-Type`,
		`Host`,
	}
	SignedHeaders := c.genSignedHeaders(signedHeaders)
	canonicalRequest += SignedHeaders + "\n"

	canonicalRequest += c.hash16Data(body)

	//log.Printf("canonicalRequest: %s", canonicalRequest)

	hashedCanonicalRequest := c.hash16Data([]byte(canonicalRequest))

	stringToSign := "SDK-HMAC-SHA256" + "\n" + xdate + "\n" + hashedCanonicalRequest

	signGen := hmac.New(sha256.New, []byte(c.sk))
	signGen.Write([]byte(stringToSign))
	signString := fmt.Sprintf("%x", signGen.Sum(nil))

	Authorization := fmt.Sprintf(" SDK-HMAC-SHA256 Access=%s, SignedHeaders=%s, Signature=%s", c.ak, SignedHeaders, signString)
	//log.Printf("Authorization: %s", Authorization)

	requestURL := fmt.Sprintf("https://%s%s", c.endpoint, resPath)
	if queryString != "" {
		requestURL += "?" + queryString
	}
	//moduleLogger.Debugf("requestURL: %s", requestURL)

	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, _ := http.NewRequest(method, requestURL, bodyReader)
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(body)))
	req.Header.Add("X-Sdk-Date", xdate)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Host", c.endpoint)
	req.Header.Add("Authorization", Authorization)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("fail for url: %s, %s", requestURL, err)
		return nil, err
	}
	defer response.Body.Close()
	respData, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		err = fmt.Errorf("%s", string(respData))
		return nil, err
	}
	return respData, err
}

func (c *HWClient) genCanonicalQuery(querys map[string]string) string {
	if querys == nil {
		return ""
	}
	canonicalQueryString := ""
	keys := []string{}
	for k := range querys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		qk := url.QueryEscape(k)
		qv := url.QueryEscape(querys[k])
		qv = strings.Replace(qv, "+", `%20`, -1)
		q := qk + "=" + qv
		if canonicalQueryString != "" {
			canonicalQueryString += "&"
		}
		canonicalQueryString += q
	}
	return canonicalQueryString
}

func (c *HWClient) genCanonicalHeaders(headers map[string]string) string {
	canonicalHeaders := ""
	headersLower := map[string]string{}
	for k, v := range headers {
		headersLower[strings.ToLower(k)] = strings.TrimSpace(v)
	}

	var keys []string
	for k := range headersLower {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		head := k + ":" + headersLower[k] + "\n"
		canonicalHeaders += head
	}
	return canonicalHeaders + "\n"
}

func (c *HWClient) genSignedHeaders(signedHeaders []string) string {
	var keys []string
	for _, k := range signedHeaders {
		keys = append(keys, strings.ToLower(k))
	}
	sort.Strings(keys)

	result := ""
	for _, k := range keys {
		if result != "" {
			result += ";"
		}
		result += k
	}
	return result
}

func (c *HWClient) hash16Data(body []byte) string {
	hashString := ""
	if body == nil {
		body = []byte("")
	}
	sha := sha256.New()
	sha.Write(body)
	hashString = fmt.Sprintf("%x", sha.Sum(nil))
	return hashString
}
