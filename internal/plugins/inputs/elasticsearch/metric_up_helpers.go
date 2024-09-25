// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package elasticsearch

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
	l.Debugf("set up metric for %q to 1", server)
	ipt.serverInfoMutex.Lock()
	ipt.upStates[server] = 1
	ipt.serverInfoMutex.Unlock()
}

func (ipt *Input) setErrUpState(server string) {
	l.Debugf("set up metric for %q to 0", server)
	ipt.serverInfoMutex.Lock()
	ipt.upStates[server] = 0
	ipt.serverInfoMutex.Unlock()
}

func (ipt *Input) getUpState(server string) int {
	ipt.serverInfoMutex.Lock()
	defer ipt.serverInfoMutex.Unlock()

	if x, ok := ipt.upStates[server]; !ok {
		l.Errorf("up status for server %q not set, should not been here", server)
		return -1 // not set yet, should not been here
	} else {
		return x
	}
}

func (ipt *Input) getUpJob() string {
	return inputName
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
		l.Errorf("failed to split host and port: %s", err)
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
		"up": ipt.getUpState(server),
	}
	m := &upMeasurement{
		name:     "collector",
		tags:     tags,
		fields:   fields,
		election: ipt.Election,
	}
	l.Debugf("build up %s points:%s", inputName, m.Point().LineProto())
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
		if err := ipt.feeder.FeedV2(point.Metric, pts,
			dkio.WithCollectCost(time.Since(time.Now())),
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
}
