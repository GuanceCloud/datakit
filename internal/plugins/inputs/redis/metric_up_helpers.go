// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var upFeedSource = dkio.FeedSource(inputName, "up")

func (ipt *Input) feedUpMetric() {
	pts := make([]*point.Point, 0)

	for _, inst := range ipt.instances {
		for _, n := range inst.nodes() {
			upState, ok := inst.nodeUpStates[n.addr]
			if !ok {
				continue
			}
			var kvs point.KVs
			kvs = kvs.AddTag("job", inputName).
				AddTag("instance", n.addr).
				Set("up", upState)

			for k, v := range ipt.Tags {
				kvs.AddTag(k, v)
			}

			opts := append(point.DefaultMetricOptions(), point.WithTime(ipt.ptsTime))
			pts = append(pts, point.NewPoint(inputs.CollectorUpMeasurement, kvs, opts...))
		}
	}

	if err := ipt.feeder.Feed(point.Metric, pts,
		dkio.WithElection(ipt.Election),
		dkio.WithSource(upFeedSource),
	); err != nil {
		l.Warnf("feed : %s, ignored", err)
	}
}
