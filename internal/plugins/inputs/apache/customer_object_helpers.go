// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package apache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	gcPoint "github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) setIptCOStatus() {
	ipt.CollectCoStatus = "OK"
}

func (ipt *Input) setIptErrCOStatus() {
	ipt.CollectCoStatus = "NotOK"
}

func (ipt *Input) setIptLastCOInfo(m *customerObjectMeasurement) {
	ipt.LastCustomerObject = m
}

func (ipt *Input) setInptErrCOMsg(s string) {
	ipt.CollectCoErrMsg = s
}

func (ipt *Input) setIptLastCOInfoByErr() {
	ipt.LastCustomerObject.tags["reason"] = ipt.CollectCoErrMsg
	ipt.LastCustomerObject.tags["col_co_status"] = ipt.CollectCoStatus
}

func (ipt *Input) getCoPointByColErr() []*gcPoint.Point {
	ms := []inputs.MeasurementV2{}
	if ipt.LastCustomerObject == nil {
		parsedURL, err := url.Parse(ipt.URL)
		if err != nil {
			return []*gcPoint.Point{}
		}
		host := parsedURL.Hostname()
		portStr := parsedURL.Port()
		if portStr == "" {
			if parsedURL.Scheme == "http" {
				portStr = "80"
			} else if parsedURL.Scheme == "https" {
				portStr = "443"
			}
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			l.Errorf("atoi port err: %s", err.Error())
		}
		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%d", host, port),
		}
		tags := map[string]string{
			"reason":        ipt.CollectCoErrMsg,
			"name":          fmt.Sprintf("%s-%s:%d", inputName, host, port),
			"host":          host,
			"ip":            fmt.Sprintf("%s:%d", host, port),
			"col_co_status": ipt.CollectCoStatus,
		}
		m := &customerObjectMeasurement{
			name:     "web_server",
			tags:     tags,
			fields:   fields,
			election: ipt.Election,
		}
		ipt.setIptLastCOInfo(m)
		ms = append(ms, m)
	} else {
		ipt.setIptLastCOInfoByErr()
		ms = append(ms, ipt.LastCustomerObject)
	}
	pts := getPointsFromMeasurement(ms)

	return pts
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*gcPoint.Point {
	pts := []*gcPoint.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func (ipt *Input) getApacheVersionAndUptime() error {
	// 发送请求到 Apache mod_status 页面
	req, err := http.NewRequest("GET", ipt.URL, nil)
	if err != nil {
		return fmt.Errorf("error on new request to %s: %w", ipt.URL, err)
	}

	if len(ipt.Username) != 0 && len(ipt.Password) != 0 {
		req.SetBasicAuth(ipt.Username, ipt.Password)
	}

	resp, err := ipt.client.Do(req)
	if err != nil {
		return fmt.Errorf("error on request to %s: %w", ipt.URL, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status %s", ipt.URL, resp.Status)
	}

	// 读取响应体内容
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	bodyText := string(bodyBytes)

	// 解析 Apache 版本信息
	versionLinePrefix := "ServerVersion: Apache/"
	startIndex := strings.Index(bodyText, versionLinePrefix)
	if startIndex == -1 {
		return fmt.Errorf("version line not found in mod_status response")
	}

	startIndex += len(versionLinePrefix)
	endIndex := strings.Index(bodyText[startIndex:], " ")
	if endIndex == -1 {
		return fmt.Errorf("end of version line not found")
	}

	ipt.Version = bodyText[startIndex : startIndex+endIndex]

	// 解析运行时长
	uptimeLinePrefix := "ServerUptimeSeconds:"
	startIndex = strings.Index(bodyText, uptimeLinePrefix)
	if startIndex == -1 {
		return fmt.Errorf("uptime line not found in mod_status response")
	}

	startIndex += len(uptimeLinePrefix)
	endIndex = strings.Index(bodyText[startIndex:], "\n")
	if endIndex == -1 {
		return fmt.Errorf("end of uptime line not found")
	}

	uptimeLine := strings.TrimSpace(bodyText[startIndex : startIndex+endIndex])
	uptimeSeconds, err := strconv.Atoi(uptimeLine)
	if err != nil {
		return fmt.Errorf("error parsing uptime seconds: %w", err)
	}

	ipt.Uptime = uptimeSeconds

	return nil
}

func (ipt *Input) collectCustomerObjectMeasurement() ([]*gcPoint.Point, error) {
	ipt.setIptCOStatus()
	ms := []inputs.MeasurementV2{}
	parsedURL, err := url.Parse(ipt.URL)
	if err != nil {
		return []*gcPoint.Point{}, nil
	}
	host := parsedURL.Hostname()
	portStr := parsedURL.Port()
	if portStr == "" {
		if parsedURL.Scheme == "http" {
			portStr = "80"
		} else if parsedURL.Scheme == "https" {
			portStr = "443"
		}
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		l.Errorf("atoi port err: %s", err.Error())
	}
	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", host, port),
		"version":      ipt.Version,
		"uptime":       fmt.Sprintf("%d", ipt.Uptime),
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("%s-%s:%d", inputName, host, port),
		"host":          host,
		"ip":            fmt.Sprintf("%s:%d", host, port),
		"col_co_status": ipt.CollectCoStatus,
	}
	m := &customerObjectMeasurement{
		name:     "web_server",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	ipt.setIptLastCOInfo(m)
	ms = append(ms, m)
	if len(ms) > 0 {
		pts := getPointsFromMeasurement(ms)
		return pts, nil
	}
	return []*gcPoint.Point{}, nil
}

func (ipt *Input) FeedCoPts() {
	err := ipt.getApacheVersionAndUptime()
	if err != nil {
		ipt.FeedCoByErr(err)
		return
	}
	pts, _ := ipt.collectCustomerObjectMeasurement()
	if err := ipt.feeder.Feed(gcPoint.CustomObject, pts,
		dkio.WithElection(ipt.Election),
		dkio.WithSource(customObjectFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
		)
		l.Errorf("feed : %s", err)
	}
}

func (ipt *Input) FeedCoByErr(err error) {
	ipt.setInptErrCOMsg(err.Error())
	ipt.setIptErrCOStatus()
	pts := ipt.getCoPointByColErr()
	if err := ipt.feeder.Feed(gcPoint.CustomObject, pts,
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
