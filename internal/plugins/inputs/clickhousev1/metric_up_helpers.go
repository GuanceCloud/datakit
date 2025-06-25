// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package clickhousev1

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) setUpState(server string) {
	ipt.upStates[server] = 1
}

func (ipt *Input) setErrUpState(server string) {
	ipt.upStates[server] = 0
}

func (ipt *Input) getUpJob() string {
	return source
}

func (ipt *Input) getUpInstance(server string) string {
	uu, _ := url.Parse(server)
	h, p, err := net.SplitHostPort(uu.Host)
	var host string
	var port int
	if err == nil {
		host = h
		port, _ = strconv.Atoi(p)
	} else {
		host = uu.Host
		ipt.l.Errorf("failed to split host and port: %s", err)
	}
	ipPort := fmt.Sprintf("%s:%d", host, port)
	return ipPort
}

func (ipt *Input) buildUpPoints(server string) ([]*point.Point, error) {
	ms := []inputs.MeasurementV2{}
	tags := map[string]string{
		"job":      ipt.getUpJob(),
		"instance": ipt.getUpInstance(server),
	}
	fields := map[string]interface{}{
		"up": ipt.upStates[server],
	}
	m := &inputs.UpMeasurement{
		Name:     inputs.CollectorUpMeasurement,
		Tags:     tags,
		Fields:   fields,
		Election: ipt.Election,
	}
	ipt.l.Debugf("build up %s points:%s", inputName, m.Point().LineProto())
	ms = append(ms, m)
	if len(ms) > 0 {
		pts := getPointsFromMeasurement2(ms)
		for k, v := range ipt.Tags {
			for _, pt := range pts {
				pt.AddTag(k, v)
			}
		}
		return pts, nil
	}

	return []*point.Point{}, nil
}

func getPointsFromMeasurement2(ms []inputs.MeasurementV2) []*point.Point {
	pts := []*point.Point{}
	for _, m := range ms {
		pts = append(pts, m.Point())
	}

	return pts
}

func (ipt *Input) FeedUpMetric(server string) {
	pts, _ := ipt.buildUpPoints(server)
	if len(pts) > 0 {
		if err := ipt.feeder.Feed(point.Metric, pts,
			dkio.WithCollectCost(time.Since(ipt.start)),
			dkio.WithElection(ipt.Election),
			dkio.WithSource(inputName)); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
			l.Errorf("feed measurement: %s", err)
		}
	}
}
