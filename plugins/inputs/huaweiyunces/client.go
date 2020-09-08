package huaweiyunces

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type metricItem struct {
	Namespace  string       `json:"namespace"`
	MetricName string       `json:"metric_name"`
	Dimensions []*Dimension `json:"dimensions"`
}

type batchReq struct {
	Period  string        `json:"period"`
	Filter  string        `json:"filter"`
	From    int64         `json:"from"`
	To      int64         `json:"to"`
	Metrics []*metricItem `json:"metrics"`
}

type hwClient struct {
	ak        string
	sk        string
	endpoint  string
	projectid string
}

type dataPoint struct {
	value float64
	tm    int64
	unit  string
}

type metricResult struct {
	metricName string
	datapoints []*dataPoint
}

func (m *metricResult) String() string {

	fs := ""
	for _, dp := range m.datapoints {
		s := fmt.Sprintf("%s %s %v%s", m.metricName, time.Unix(dp.tm/1000, 0), dp.value, dp.unit)
		fs += s + "\n"
	}
	return fs
}

type batchMetricResultItem struct {
	namespace  string
	metricName string
	unit       string
	datapoints []*dataPoint
}

type batchMetricResult struct {
	results []*batchMetricResultItem
}

func (m *batchMetricResultItem) String() string {

	fs := ""
	for _, dp := range m.datapoints {
		s := fmt.Sprintf("%s(%s) %s %v%s", m.metricName, m.namespace, time.Unix(dp.tm/1000, 0), dp.value, m.unit)
		fs += s + "\n"
	}
	return fs
}

func newHWClient(ak, sk, endpoint, projectid string) *hwClient {
	return &hwClient{
		ak:        ak,
		sk:        sk,
		endpoint:  endpoint,
		projectid: projectid,
	}
}

func (c *hwClient) getBatchMetricResourcePath() string {
	return fmt.Sprintf("/V1.0/%s/batch-query-metric-data", c.projectid)
}

func (c *hwClient) getMetricResourcePath() string {
	return fmt.Sprintf("/V1.0/%s/metric-data", c.projectid)
}

func (c *hwClient) genCanonicalQuery(querys map[string]string) string {
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

func (c *hwClient) genCanonicalHeaders(headers map[string]string) string {
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

func (c *hwClient) genSignedHeaders(signedHeaders []string) string {
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

func (c *hwClient) hash16Data(body []byte) string {
	hashString := ""
	if body == nil {
		body = []byte("")
	}
	sha := sha256.New()
	sha.Write(body)
	hashString = fmt.Sprintf("%x", sha.Sum(nil))
	return hashString
}

func (c *hwClient) request(method string, resPath string, querys map[string]string, body []byte) ([]byte, error) {

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

func (c *hwClient) getMetric(namespace, metricname string, filter string, period int, from, to int64, dims []*Dimension) (*metricResult, error) {
	querys := map[string]string{
		"namespace":   namespace,
		"metric_name": metricname,
		"from":        fmt.Sprintf("%d", from),
		"to":          fmt.Sprintf("%d", to),
		"period":      fmt.Sprintf("%d", period),
		"filter":      filter,
	}
	for i, d := range dims {
		querys[fmt.Sprintf("dim.%d", i)] = fmt.Sprintf("%s,%s", d.Name, d.Value)
	}
	resp, err := c.request("GET", c.getMetricResourcePath(), querys, nil)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return nil, err
	}
	//moduleLogger.Debugf("%s", string(resp))

	return parseMetricResponse(resp, filter), nil
}

func (c *hwClient) batchMetrics(b *batchReq) ([]byte, error) {
	data, err := json.Marshal(b)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return nil, err
	}
	resp, err := c.request("POST", c.getBatchMetricResourcePath(), nil, data)
	if err != nil {
		moduleLogger.Errorf("%s", err)
		return nil, err
	}
	return resp, nil
}

func parseMetricResponse(resp []byte, filter string) *metricResult {
	var resJSON map[string]interface{}
	err := json.Unmarshal(resp, &resJSON)
	if err != nil {
		moduleLogger.Errorf("fail to unmarshal, %s", err)
		return nil
	}
	var result metricResult
	if s, ok := resJSON["metric_name"].(string); ok {
		result.metricName = s
	}
	dps := resJSON["datapoints"]
	if dpArr, ok := dps.([]interface{}); ok {
		for _, dp := range dpArr {
			var datapoint dataPoint
			if dpMap, ok := dp.(map[string]interface{}); ok {
				if v, ok := dpMap[filter].(float64); ok {
					datapoint.value = v
				}
				if t, ok := dpMap["timestamp"].(float64); ok {
					datapoint.tm = int64(t)
				}
				if u, ok := dpMap["unit"].(string); ok {
					datapoint.unit = u
				}
			}
			result.datapoints = append(result.datapoints, &datapoint)
		}

	}

	return &result
}

func parseBatchResponse(resp []byte, filter string) *batchMetricResult {
	var results map[string]interface{}
	err := json.Unmarshal(resp, &results)
	if err != nil {
		moduleLogger.Errorf("fail to unmarshal, %s", err)
		return nil
	}
	metrics := results["metrics"]
	//log.Printf("%v", reflect.TypeOf(metrics))
	var batchResult batchMetricResult
	if metricArr, ok := metrics.([]interface{}); ok {
		for _, item := range metricArr {
			var resItem batchMetricResultItem
			if mitem, ok := item.(map[string]interface{}); ok {
				if s, ok := mitem["namespace"].(string); ok {
					resItem.namespace = s
				}
				if s, ok := mitem["metric_name"].(string); ok {
					resItem.metricName = s
				}
				if s, ok := mitem["unit"].(string); ok {
					resItem.unit = s
				}
				if dpArr, ok := mitem["datapoints"].([]interface{}); ok {
					for _, dp := range dpArr {
						var datapoint dataPoint
						if dpMap, ok := dp.(map[string]interface{}); ok {
							if v, ok := dpMap[filter].(float64); ok {
								datapoint.value = v
							}
							if t, ok := dpMap["timestamp"].(float64); ok {
								datapoint.tm = int64(t)
							}
						}
						resItem.datapoints = append(resItem.datapoints, &datapoint)
					}
				}
			}
			batchResult.results = append(batchResult.results, &resItem)
		}

	}

	return &batchResult
}
