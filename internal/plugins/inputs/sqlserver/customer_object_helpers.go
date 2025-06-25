// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"context"
	"fmt"
	"net"
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
	ms := []inputs.MeasurementV2{}
	if ipt.LastCustomerObject == nil {
		ip, port, err := net.SplitHostPort(ipt.Host)
		if err != nil {
			l.Errorf("Failed to parse host %s: %v", ipt.Host, err)
		}
		portInt, err := strconv.Atoi(port)
		if err != nil {
			l.Errorf("Failed to convert port %s to int: %v", port, err)
		}
		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%d", ip, portInt),
		}
		tags := map[string]string{
			"reason":        ipt.CollectCoErrMsg,
			"name":          fmt.Sprintf("%s-%s:%d", inputName, ip, portInt),
			"host":          ip,
			"ip":            fmt.Sprintf("%s:%d", ip, portInt),
			"col_co_status": ipt.CollectCoStatus,
		}
		m := &customerObjectMeasurement{
			name:     "database",
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

func (ipt *Input) getVersionAndUptime() error {
	ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
	defer cancel()

	versionQuery := "SELECT @@VERSION"
	var version string
	err := ipt.db.QueryRowContext(ctx, versionQuery).Scan(&version)
	if err != nil {
		return fmt.Errorf("failed to get SQL Server version: %w", err)
	}
	ipt.Version = version

	// 查询 SQL Server 启动时间
	uptimeQuery := `
		SELECT
			DATEDIFF(SECOND, sqlserver_start_time, GETDATE())
		FROM
			sys.dm_os_sys_info
	`
	var uptime int
	err = ipt.db.QueryRowContext(ctx, uptimeQuery).Scan(&uptime)
	if err != nil {
		return fmt.Errorf("failed to get SQL Server uptime: %w", err)
	}
	ipt.Uptime = uptime

	return nil
}

func (ipt *Input) collectCustomerObjectMeasurement() ([]*gcPoint.Point, error) {
	ipt.setIptCOStatus()
	ms := []inputs.MeasurementV2{}

	ip, port, err := net.SplitHostPort(ipt.Host)
	if err != nil {
		ip = ipt.Host
		port = "1433"
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		l.Errorf("Failed to convert port %s to int: %v", port, err)
	}
	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", ip, portInt),
		"version":      ipt.Version,
		"uptime":       ipt.Uptime,
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("%s-%s:%d", inputName, ip, portInt),
		"host":          ip,
		"ip":            fmt.Sprintf("%s:%d", ip, portInt),
		"col_co_status": ipt.CollectCoStatus,
	}
	m := &customerObjectMeasurement{
		name:     "database",
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
		l.Errorf("getVersionAndUptime error: %v", err)
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
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}

func (ipt *Input) FeedCoByErr(err error) {
	ipt.setIptErrCOMsg(err.Error())
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
