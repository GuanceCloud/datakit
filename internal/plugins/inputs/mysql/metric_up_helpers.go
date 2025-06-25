// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
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
	return ipt.Addr
}

func (ipt *Input) buildUpPoints() ([]*point.Point, error) {
	var pts []*point.Point
	opts := ipt.getKVsOpts()
	kvs := ipt.getKVs()

	kvs = kvs.AddTag("job", ipt.getUpJob())
	kvs = kvs.AddTag("instance", ipt.getUpInstance())

	kvs = kvs.Add("up", ipt.UpState, false, true)

	if kvs.FieldCount() > 0 {
		pts = append(pts, point.NewPointV2("collector", kvs, opts...))
	}

	return pts, nil
}

func (ipt *Input) FeedUpMetric() {
	pts, _ := ipt.buildUpPoints()
	if len(pts) > 0 {
		l.Debug("feed up metric")
		if err := ipt.feeder.Feed(point.Metric, pts,
			dkio.WithElection(ipt.Election),
			dkio.WithSource(inputName),
		); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Metric),
			)
			l.Errorf("feed : %s", err)
		}
	}
}
