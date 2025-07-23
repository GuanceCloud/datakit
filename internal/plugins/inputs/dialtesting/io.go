// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dialtesting

import (
	"time"

	pt "github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

func (d *dialer) pointsFeed(urlStr string) {
	d.seqNumber++
	startTime := time.Now()
	tags, fields := d.task.GetResults()

	if status, ok := tags["status"]; ok {
		taskCheckCostSummary.WithLabelValues(d.regionName, d.class, status).Observe(float64(time.Since(startTime)) / float64(time.Second))
	}

	for k, v := range d.tags {
		if d.measurementInfo != nil && d.measurementInfo.Tags != nil {
			if _, ok := d.measurementInfo.Tags[k]; !ok {
				continue
			}
		}

		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Debugf("ignore dialer tag %s: %s", k, v)
		}
	}

	fields["seq_number"] = d.seqNumber
	tags["datakit_version"] = datakit.Version
	tags["node_name"] = d.regionName

	// df tags
	for k, v := range d.dfTags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Debugf("ignore df tag %s: %s", k, v)
		}
	}

	// custom tags
	for k, v := range d.ipt.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		} else {
			l.Warnf("duplicate tag, ignore custom tag %s", k)
		}
	}

	opt := append(pt.DefaultLoggingOptions(), pt.WithTime(d.dialingTime))
	data := pt.NewPointV2(d.task.MetricName(),
		append(pt.NewTags(tags), pt.NewKVs(fields)...), opt...)

	dialWorker.addPoints(&jobData{
		url:        urlStr,
		pt:         data,
		regionName: d.regionName,
		class:      d.class,
	})
}
