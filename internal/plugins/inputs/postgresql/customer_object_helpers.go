// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

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
		uu, err := url.Parse(ipt.Address)
		if err != nil {
			l.Errorf("Failed to parse address %s: %v", ipt.Address, err)
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
			port = 5432 // 默认 PostgreSQL 端口
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
			name:   "database",
			tags:   tags,
			fields: fields,
			ipt:    ipt,
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

func (ipt *Input) getUptime() error {
	rows, err := ipt.service.Query("SELECT pg_postmaster_start_time();")
	if err != nil {
		return fmt.Errorf("failed to query PostgreSQL start time: %w", err)
	}
	defer rows.Close()

	var startTime time.Time

	if rows.Next() {
		if err := rows.Scan(&startTime); err != nil {
			return fmt.Errorf("failed to scan PostgreSQL start time: %w", err)
		}
	}

	uptime := int(time.Since(startTime).Seconds())
	ipt.Uptime = uptime

	return nil
}

func (ipt *Input) collectCustomerObjectMeasurement() ([]*gcPoint.Point, error) {
	ipt.setIptCOStatus()
	ms := []inputs.MeasurementV2{}

	uu, err := url.Parse(ipt.Address)
	if err != nil {
		l.Errorf("Failed to parse address %s: %v", ipt.Address, err)
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
		port = 5432 // 默认 PostgreSQL 端口
	}

	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", host, port),
		"version":      ipt.version.String(),
		"uptime":       fmt.Sprintf("%d", ipt.Uptime),
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("%s-%s:%d", inputName, host, port),
		"host":          host, // 主机名
		"ip":            fmt.Sprintf("%s:%d", host, port),
		"col_co_status": ipt.CollectCoStatus,
	}
	m := &customerObjectMeasurement{
		name:   "database",
		tags:   tags,
		fields: fields,
		ipt:    ipt,
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
	pts, _ := ipt.collectCustomerObjectMeasurement()
	if err := ipt.feeder.FeedV2(gcPoint.CustomObject, pts,
		dkio.WithCollectCost(time.Since(time.Now())),
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
	if err := ipt.feeder.FeedV2(gcPoint.CustomObject,
		pts,
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
