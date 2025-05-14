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

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) setUpState() {
	ipt.UpState = 1
}

func (ipt *Input) setErrUpState() {
	ipt.UpState = 0
}

func (ipt *Input) getUpJob() string {
	return inputName
}

func (ipt *Input) getUpInstance() string {
	uu, err := url.Parse(ipt.Address)
	if err != nil {
		l.Errorf("Failed to parse address %s: %v", ipt.Address, err)
		return ""
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
	ipPort := fmt.Sprintf("%s:%d", host, port)
	return ipPort
}

func (ipt *Input) buildUpPoints() *point.Point {
	tags := map[string]string{
		"job":      ipt.getUpJob(),
		"instance": ipt.getUpInstance(),
	}
	fields := map[string]interface{}{
		"up": ipt.UpState,
	}
	m := &inputs.UpMeasurement{
		Name:     inputs.CollectorUpMeasurement,
		Tags:     tags,
		Fields:   fields,
		Election: ipt.Election,
	}

	return m.Point()
}

func (ipt *Input) FeedUpMetric() {
	pt := ipt.buildUpPoints()

	l.Debug("feed up metric")

	if err := ipt.feeder.FeedV2(point.Metric, []*point.Point{pt},
		dkio.WithElection(ipt.Election),
		dkio.WithInputName(inputName),
	); err != nil {
		ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Metric),
		)
		l.Errorf("feed : %s", err)
	}
}
