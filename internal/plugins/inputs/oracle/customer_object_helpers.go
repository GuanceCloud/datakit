// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"fmt"
	"regexp"
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
		fields := map[string]interface{}{
			"display_name": fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
		}
		tags := map[string]string{
			"reason":        ipt.CollectCoErrMsg,
			"name":          fmt.Sprintf("%s-%s:%d", inputName, ipt.Host, ipt.Port),
			"host":          ipt.Host,
			"ip":            fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
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

func (ipt *Input) getOracleVersion() {
	if ipt.db == nil {
		l.Errorf("Database connection is nil")
		return
	}

	versionQuery := "SELECT BANNER FROM v$version WHERE banner LIKE 'Oracle%'"
	if err := ipt.db.Get(&ipt.fullVersion, versionQuery); err != nil {
		l.Errorf("Failed to get Oracle version: %s", err.Error())
		return
	} else if ipt.mainVersion == "" {
		if x := oracleVersionRe.FindStringSubmatch(ipt.fullVersion); len(x) > 1 {
			ipt.mainVersion = strings.Split(x[1], ".")[0]
		}
	}
}

func (ipt *Input) getOracleUptime() {
	if ipt.db == nil {
		l.Errorf("Database connection is nil")
		return
	}

	uptimeQuery := "SELECT (SYSDATE - STARTUP_TIME) * 86400 AS uptime_seconds FROM v$instance"

	// 获取 Oracle 启动时间
	var uptimeSeconds float64
	if err := ipt.db.Get(&uptimeSeconds, uptimeQuery); err != nil {
		l.Errorf("Failed to get Oracle uptime: %s", err.Error())

		ipt.setErrUpState()
		return
	}
	ipt.Uptime = int(uptimeSeconds)

	l.Debugf("Oracle Version: %s, Uptime: %d seconds", ipt.fullVersion, ipt.Uptime)
}

var oracleVersionRe = regexp.MustCompile(`Release\s+(\d+\.\d+\.\d+\.\d+\.\d+)`)

func (ipt *Input) collectCustomerObjectMeasurement() ([]*gcPoint.Point, error) {
	ipt.setIptCOStatus()
	ms := []inputs.MeasurementV2{}

	fields := map[string]interface{}{
		"display_name": fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
		"version":      ipt.fullVersion,
		"uptime":       fmt.Sprintf("%d", ipt.Uptime),
	}
	tags := map[string]string{
		"name":          fmt.Sprintf("%s-%s:%d", inputName, ipt.Host, ipt.Port),
		"host":          ipt.Host,
		"ip":            fmt.Sprintf("%s:%d", ipt.Host, ipt.Port),
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
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(gcPoint.CustomObject),
		)
		l.Errorf("feed : %s", err)
	}
}
