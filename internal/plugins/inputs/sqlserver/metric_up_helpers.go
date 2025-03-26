// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"fmt"
	"strings"
	"time"

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
	port := "1433"
	host := ipt.Host
	parts := strings.Split(ipt.Host, ":")
	if len(parts) == 2 {
		host = parts[0]
		port = parts[1]
	}

	return fmt.Sprintf("%s:%s", host, port)
}

func (ipt *Input) buildUpPoints() ([]*point.Point, error) {
	ms := []inputs.MeasurementV2{}
	tags := map[string]string{
		"job":      ipt.getUpJob(),
		"instance": ipt.getUpInstance(),
	}
	fields := map[string]interface{}{
		"up": ipt.UpState,
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

func (ipt *Input) FeedUpMetric() {
	pts, _ := ipt.buildUpPoints()
	if len(pts) > 0 {
		l.Debug("feed up metric")
		if err := ipt.feeder.FeedV2(point.Metric, pts,
			dkio.WithCollectCost(time.Since(ipt.start)),
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

func (ipt *Input) FeedErrUpMetric() {
	ipt.setErrUpState()
	pts, _ := ipt.buildUpPoints()
	if len(pts) > 0 {
		l.Debug("feed up metric")
		if err := ipt.feeder.FeedV2(point.Metric, pts,
			dkio.WithCollectCost(time.Since(ipt.start)),
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
