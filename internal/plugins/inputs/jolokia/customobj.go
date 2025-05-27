// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package jolokia

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type customObject struct {
	name     string
	tags     map[string]string
	fields   map[string]interface{}
	election bool
}

func (m *customObject) Point() *point.Point {
	opts := point.DefaultObjectOptions()

	if m.election {
		opts = append(opts, point.WithExtraTags(datakit.GlobalElectionTags()))
	}

	return point.NewPointV2(m.name,
		append(point.NewTags(m.tags), point.NewKVs(m.fields)...),
		opts...)
}

func (j *JolokiaAgent) collectCustomerObjectMeasurement(client *jclient) ([]*point.Point, error) {
	var CoPts []*point.Point
	uu, _ := url.Parse(client.url)
	h, _, err := net.SplitHostPort(uu.Host)
	var host string
	if err == nil {
		host = h
	} else {
		j.L.Errorf("failed to split host and port: %s", err)
	}
	version, uptime, err := j.getKafkaVersionAndUptime(client)
	if err != nil {
		j.L.Errorf("failed to get kafka version and uptime: %s", err)
		return []*point.Point{}, nil
	}

	j.L.Debugf("kafka version:%s,uptime:%d", version, uptime)

	fields := map[string]interface{}{
		"display_name": host,
		"version":      version,
		"uptime":       fmt.Sprintf("%d", int(uptime)),
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("%s-%s", "kafka", host),
		"host":          host,
		"ip":            host,
		"col_co_status": "OK",
	}
	Copt := &customObject{
		name:     "mq",
		tags:     tags,
		fields:   fields,
		election: j.Election,
	}
	j.L.Debugf("pts: %s", Copt.Point().LineProto())
	CoPts = append(CoPts, Copt.Point())
	if len(CoPts) > 0 {
		j.L.Debugf("pts: %s", CoPts[0].LineProto())
		return CoPts, nil
	}
	return []*point.Point{}, nil
}

func (j *JolokiaAgent) getKafkaVersionAndUptime(client *jclient) (version string, uptimeSeconds int64, err error) {
	requestURL, err := formatReadURL(client.url, client.config.username, client.config.password)
	if err != nil {
		return "", 0, fmt.Errorf("failed to format read URL: %w", err)
	}

	// 添加 MBean 请求
	mbeanURL := buildURL(requestURL, "/kafka.server:type=app-info")
	j.L.Debugf("getVersionAndUptime mbeanURL: %s", mbeanURL)

	req, err := http.NewRequest("GET", mbeanURL, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(client.config.username, client.config.password)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			j.L.Errorf("Failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("response has status code %d (%s), expected %d (%s)",
			resp.StatusCode, http.StatusText(resp.StatusCode), http.StatusOK, http.StatusText(http.StatusOK))
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析 JSON 响应
	var jResponse struct {
		Request struct {
			MBean string `json:"mbean"`
		} `json:"request"`
		Value struct {
			StartTimeMs int64  `json:"start-time-ms"`
			CommitId    string `json:"commit-id"` //nolint:stylecheck
			Version     string `json:"version"`
		} `json:"value"`
	}

	err = json.Unmarshal(responseBody, &jResponse)
	if err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	version = jResponse.Value.Version
	startTime := jResponse.Value.StartTimeMs
	uptimeMillis := time.Now().UnixMilli() - startTime
	uptimeSeconds = uptimeMillis / 1000

	return version, uptimeSeconds, nil
}

func buildURL(baseURL, relativePath string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	relativePath = strings.TrimLeft(relativePath, "/")
	return baseURL + "/" + relativePath
}
