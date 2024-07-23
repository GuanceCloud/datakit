// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package prom

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (i *Input) setUpState(server string) {
	i.upStates[server] = 1
}

func (i *Input) setErrUpState(server string) {
	i.upStates[server] = 0
}

func (i *Input) getUpJob() string {
	return i.Source
}

func (i *Input) getUpInstance(server string) string {
	uu, _ := url.Parse(server)
	h, p, err := net.SplitHostPort(uu.Host)
	var host string
	var port int
	if err == nil {
		host = h
		port, _ = strconv.Atoi(p)
	} else {
		host = uu.Host
		i.l.Errorf("failed to split host and port: %s", err)
	}
	ipPort := fmt.Sprintf("%s:%d", host, port)
	return ipPort
}

func (i *Input) buildUpPoints(server string) ([]*point.Point, error) {
	ms := []inputs.MeasurementV2{}
	tags := map[string]string{
		"job":      i.getUpJob(),
		"instance": i.getUpInstance(server),
	}
	fields := map[string]interface{}{
		"up": i.upStates[server],
	}
	m := &upMeasurement{
		name:     "collector",
		tags:     tags,
		fields:   fields,
		election: i.Election,
	}
	i.l.Debugf("build up %s points:%s", inputName, m.Point().LineProto())
	ms = append(ms, m)
	if len(ms) > 0 {
		pts := getPointsFromMeasurement2(ms)
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

func (i *Input) FeedUpMetric(server string) {
	pts, _ := i.buildUpPoints(server)
	if len(pts) > 0 {
		if err := i.Feeder.FeedV2(point.Metric, pts,
			dkio.WithCollectCost(time.Since(time.Now())),
			dkio.WithElection(i.Election),
			dkio.WithInputName(i.Source),
		); err != nil {
			i.Feeder.FeedLastError(err.Error(),
				dkio.WithLastErrorInput(i.Source),
				dkio.WithLastErrorCategory(point.Metric),
			)
		}
	}
}
