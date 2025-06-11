// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rabbitmq

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

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

func (ipt *Input) setIptErrCOMsg(s string) {
	ipt.CollectCoErrMsg = s
}

func (ipt *Input) setIptLastCOInfoByErr() {
	ipt.LastCustomerObject.tags["reason"] = ipt.CollectCoErrMsg
	ipt.LastCustomerObject.tags["col_co_status"] = ipt.CollectCoStatus
}

func (ipt *Input) getCoPointByColErr() []*gcPoint.Point {
	var ms []inputs.MeasurementV2
	if ipt.LastCustomerObject == nil {
		uu, err := url.Parse(ipt.URL)
		if err != nil {
			l.Errorf("Failed to parse address %s: %v", ipt.URL, err)
			return []*gcPoint.Point{}
		}
		var host string
		var port int
		h, p, err := net.SplitHostPort(uu.Host)
		if err == nil {
			host = h
			port, _ = strconv.Atoi(p)
		} else {
			host = uu.Host
			port = 15672
		}

		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%d", host, port),
		}
		tags := map[string]string{
			"reason":        ipt.CollectCoErrMsg,
			"name":          fmt.Sprintf("%s-%s:%d", inputName, host, port),
			"host":          host, // 主机名
			"ip":            fmt.Sprintf("%s:%d", host, port),
			"col_co_status": ipt.CollectCoStatus,
		}
		m := &customerObjectMeasurement{
			name:     "mq",
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

	// 从测量数据生成点
	pts := getPointsFromMeasurement(ms)
	if len(pts) == 0 {
		l.Warnf("No points generated from measurements")
	}

	return pts
}

func getPointsFromMeasurement(ms []inputs.MeasurementV2) []*gcPoint.Point {
	pts := []*gcPoint.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func (ipt *Input) getVersionAndUptime() error {
	type OverviewResponse struct {
		Version string `json:"rabbitmq_version"`
	}
	overview := &OverviewResponse{}
	err := ipt.requestJSON("/api/overview", overview)
	if err != nil {
		return err
	}
	ipt.Version = overview.Version

	// 定义解析 /api/nodes 响应体的结构体
	type NodeResponse struct {
		Uptime int `json:"uptime"`
	}

	// 发送请求并解析 JSON 响应体
	var nodes []NodeResponse
	err = ipt.requestJSON("/api/nodes", &nodes)
	if err != nil {
		return err
	}

	// 从返回的节点数组中提取第一个节点的 uptime
	if len(nodes) > 0 {
		ipt.Uptime = nodes[0].Uptime / 1000 // 将 uptime 从毫秒转换为秒
	} else {
		l.Errorf("no nodes data available")
	}

	return nil
}

func (ipt *Input) collectCustomerObjectMeasurement() ([]*gcPoint.Point, error) {
	ipt.setIptCOStatus()
	ms := []inputs.MeasurementV2{}

	uu, err := url.Parse(ipt.URL)
	if err != nil {
		l.Errorf("Failed to parse address %s: %v", ipt.URL, err)
		return []*gcPoint.Point{}, nil
	}
	var host string
	var port int
	h, p, err := net.SplitHostPort(uu.Host)
	if err == nil {
		host = h
		port, _ = strconv.Atoi(p)
	} else {
		host = uu.Host
		port = 15672
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
		name:     "mq",
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
	err := ipt.getVersionAndUptime()
	if err != nil {
		l.Errorf("Failed to get version and uptime: %v", err)
		ipt.FeedCoByErr(err)
		return
	}
	pts, _ := ipt.collectCustomerObjectMeasurement()
	if err := ipt.feeder.FeedV2(gcPoint.CustomObject, pts,
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(customObjectFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}

func (ipt *Input) FeedCoByErr(err error) {
	ipt.setIptErrCOMsg(err.Error())
	ipt.setIptErrCOStatus()
	pts := ipt.getCoPointByColErr()
	if err := ipt.feeder.FeedV2(gcPoint.CustomObject, pts,
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(customObjectFeedName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(customObjectFeedName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}
