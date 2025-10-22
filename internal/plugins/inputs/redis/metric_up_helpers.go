// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package redis

import (
	"strings"

	"github.com/GuanceCloud/cliutils/point"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func (ipt *Input) getUpInstance() string {
	if ipt.Cluster != nil { // redis cluster
		return strings.Join(ipt.Cluster.Hosts, ",")
	}
	if ipt.MasterSlave != nil {
		return strings.Join(ipt.MasterSlave.Hosts, ",")
	}
	return ipt.Host
}

func (ipt *Input) buildUpPoints() *point.Point {
	var kvs point.KVs
	kvs = kvs.AddTag("job", inputName).
		AddTag("instance", ipt.getUpInstance()).
		Set("up", ipt.upState)

	for k, v := range ipt.Tags {
		kvs.AddTag(k, v)
	}

	opts := append(point.DefaultMetricOptions(), point.WithTime(ipt.ptsTime))
	return point.NewPoint(inputs.CollectorUpMeasurement, kvs, opts...)
}

var upFeedSource = dkio.FeedSource(inputName, "up")

func (ipt *Input) feedUpMetric() {
	pt := ipt.buildUpPoints()

	if err := ipt.feeder.Feed(point.Metric, []*point.Point{pt},
		dkio.WithElection(ipt.Election),
		dkio.WithSource(upFeedSource),
	); err != nil {
		l.Warnf("feed : %s, ignored", err)
	}
}
