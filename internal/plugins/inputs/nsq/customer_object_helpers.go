// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package nsq

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"

	gcPoint "github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) setIptLastCOInfo(url string, m *customerObjectMeasurement) {
	ipt.CustomerObjectMap[url] = m
}

func (ipt *Input) fixIptLastCOInfo(url string, version string, uptime int) {
	ipt.CustomerObjectMap[url].fields["version"] = version
	ipt.CustomerObjectMap[url].fields["uptime"] = strconv.Itoa(uptime)
}

func (ipt *Input) setIptLastCOInfoByErr(url string, reason string) {
	ipt.CustomerObjectMap[url].tags["reason"] = reason
	ipt.CustomerObjectMap[url].tags["col_co_status"] = "NotOK"
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*gcPoint.Point {
	pts := []*gcPoint.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func (ipt *Input) getVersionAndUptime(url string) (string, int, error) {
	type StatsResponse struct {
		Version   string `json:"version"`
		StartTime int64  `json:"start_time"`
	}
	resp, err := ipt.httpClient.Get(url)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Errorf("error closing response body: %v", err)
		}
	}(resp.Body)

	// 解析 JSON 数据
	var stats StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	// 获取当前时间
	currentTime := time.Now().Unix()

	// 计算启动时间
	uptime := int(currentTime - stats.StartTime)

	return stats.Version, uptime, nil
}

func (ipt *Input) collectCustomerObjectMeasurement() ([]*gcPoint.Point, error) {
	ms := []inputs.MeasurementV2{}
	for u := range ipt.nsqdEndpointList {
		uu, _ := url.Parse(u)
		h, p, err := net.SplitHostPort(uu.Host)
		var host string
		var port int
		if err == nil {
			host = h
			port, _ = strconv.Atoi(p)
		} else {
			l.Errorf("failed to split host and port: %s", err)
		}
		version, uptime, err := ipt.getVersionAndUptime(u)
		if err != nil {
			l.Errorf("failed to get version and uptime: %s", err)
		}
		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%d", host, port),
			"version":      version,
			"uptime":       fmt.Sprintf("%d", uptime),
		}
		tags := map[string]string{
			"name":          fmt.Sprintf("%s-%s:%d", inputName, host, port),
			"host":          host,
			"ip":            fmt.Sprintf("%s:%d", host, port),
			"col_co_status": "OK",
		}
		m := &customerObjectMeasurement{
			name:     "mq",
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		}
		if ipt.CustomerObjectMap[u] == nil {
			ipt.setIptLastCOInfo(u, m)
		} else if err == nil {
			ipt.fixIptLastCOInfo(u, version, uptime)
		}

		if err != nil {
			ipt.setIptLastCOInfoByErr(u, err.Error())
		}
		ms = append(ms, ipt.CustomerObjectMap[u])
	}

	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		l.Debugf("pts: %v", pts)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (ipt *Input) getCoPointByColErr(u string, errStr string) []*gcPoint.Point {
	var ms []inputs.MeasurementV2
	uu, _ := url.Parse(u)
	h, p, err := net.SplitHostPort(uu.Host)
	var host string
	var port int
	if err == nil {
		host = h
		port, _ = strconv.Atoi(p)
	} else {
		l.Errorf("failed to split host and port: %s", err)
	}
	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", host, port),
	}
	tags := map[string]string{
		"reason":        errStr,
		"name":          fmt.Sprintf("%s-%s:%d", inputName, host, port),
		"host":          host,
		"ip":            fmt.Sprintf("%s:%d", host, port),
		"col_co_status": "NotOK",
	}
	m := &customerObjectMeasurement{
		name:     "mq",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	ms = append(ms, m)
	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts
	}
	return []*gcPoint.Point{}
}

func (ipt *Input) FeedCoPts() {
	pts, err := ipt.collectCustomerObjectMeasurement()
	if err != nil {
		l.Errorf("failed to collect customer object measurements: %s", err)
	}
	if err := ipt.feeder.Feed(gcPoint.CustomObject, pts,
		dkio.WithCollectCost(time.Since(time.Now())),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(customObjectFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}

func (ipt *Input) FeedCoByErr(url string, err error) {
	pts := ipt.getCoPointByColErr(url, err.Error())
	if err := ipt.feeder.Feed(gcPoint.CustomObject, pts,
		dkio.WithCollectCost(time.Since(time.Now())),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(customObjectFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}
